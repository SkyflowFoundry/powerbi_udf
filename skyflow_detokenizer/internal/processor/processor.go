// internal/processor/processor.go
package processor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"skyflow_detokenizer/internal/logging"
	"skyflow_detokenizer/internal/models"
	"sync"
	"time"
)

// TokenProcessor handles the core business logic of detokenization
type TokenProcessor struct {
	VaultURL           string       // Skyflow vault URL
	BearerToken        string       // Authentication token for Skyflow API
	BatchSize          int          // Maximum number of tokens to process in one API call
	ParallelCalls      int          // Maximum number of parallel API calls
	RequestTimeoutSecs int          // Timeout in seconds for API calls
	Client             *http.Client // HTTP client with timeout configuration
}

// NewTokenProcessor creates a new processor with the specified configuration
func NewTokenProcessor(vaultURL, bearerToken string, batchSize, parallelCalls, requestTimeoutSecs int) *TokenProcessor {
	return &TokenProcessor{
		VaultURL:           vaultURL,
		BearerToken:        bearerToken,
		BatchSize:          batchSize,
		ParallelCalls:      parallelCalls,
		RequestTimeoutSecs: requestTimeoutSecs,
		Client: &http.Client{
			Timeout: time.Second * time.Duration(requestTimeoutSecs),
		},
	}
}

// ProcessTokenBatch handles detokenization for a single batch of tokens
func (p *TokenProcessor) ProcessTokenBatch(tokens []string, redaction string) (map[string]string, error) {
	results := make(map[string]string)

	// Skip processing if no tokens provided
	if len(tokens) == 0 {
		logging.Info("No tokens to process in batch")
		return results, nil
	}

	// Log token count at info level (no sensitive data)
	logging.Info("Processing batch with %d tokens", len(tokens))
	
	// Only log actual tokens at debug level
	if len(tokens) > 0 {
		logging.Debug("First token in batch: %s", tokens[0])
	}

	// Create the request for Skyflow API
	skyflowReq := models.SkyflowRequest{
		DetokenizationParameters: make([]struct {
			Token     string `json:"token"`
			Redaction string `json:"redaction"`
		}, len(tokens)),
		ContinueOnError: true, // Enable continue on error
	}

	// Prepare the parameters for each token
	for i, token := range tokens {
		skyflowReq.DetokenizationParameters[i].Token = token
		skyflowReq.DetokenizationParameters[i].Redaction = redaction
	}

	// Marshal the request to JSON
	jsonData, err := json.Marshal(skyflowReq)
	if err != nil {
		logging.Error("Error marshaling request: %v", err)
		return results, err
	}

	// Log request details only at debug level
	logging.Debug("Sending request to Skyflow API with %d tokens", len(tokens))

	// Create new HTTP request
	req, err := http.NewRequest("POST", p.VaultURL+"/detokenize", bytes.NewBuffer(jsonData))
	if err != nil {
		logging.Error("Error creating request: %v", err)
		return results, err
	}

	// Set required headers
	req.Header.Set("Authorization", "Bearer "+p.BearerToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the request to Skyflow
	logging.Info("Sending detokenization request to Skyflow API")
	resp, err := p.Client.Do(req)
	if err != nil {
		logging.Error("Error making request: %v", err)
		return results, err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Error("Error reading response body: %v", err)
		return results, err
	}

	// Check response status - now accepting both 200 and 207 as valid responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
		logging.Error("Error response from Skyflow: %d", resp.StatusCode)
		// Only log response body at debug level as it might contain sensitive data
		logging.Debug("Error response body: %s", string(body))
		return results, err
	}

	// Log success at info level (no sensitive data)
	logging.Info("Received successful response from Skyflow API with status code %d", resp.StatusCode)

	// Parse the response
	var skyflowResp models.SkyflowResponse
	if err := json.Unmarshal(body, &skyflowResp); err != nil {
		logging.Error("Error unmarshaling response: %v", err)
		return results, err
	}

	// Log the response for debugging (only at debug level)
	logging.Debug("Received response from Skyflow with %d records", len(skyflowResp.Records))

	// Map the results - handling both successful and error cases
	for _, record := range skyflowResp.Records {
		if record.Error != nil && *record.Error == "Token Not Found" {
			results[record.Token] = "Token Not Found"
			logging.Debug("Token not found: %s", record.Token)
		} else if record.Value != "" {
			results[record.Token] = record.Value
			// Don't log the actual value, even at debug level, just log that we got a value
			logging.Debug("Received value for token: %s", record.Token)
		} else {
			// For any other error case or empty value
			results[record.Token] = "Token Not Found"
			logging.Debug("No value or error for token: %s", record.Token)
		}
	}

	logging.Info("Successfully processed %d tokens", len(results))
	return results, nil
}

// ProcessTokens handles the detokenization of multiple tokens with batching and parallelization
func (p *TokenProcessor) ProcessTokens(tokens []string, redaction string) map[string]string {
	// Master results map
	results := make(map[string]string)
	resultsMutex := &sync.Mutex{}

	// Skip processing if no tokens provided
	if len(tokens) == 0 {
		logging.Info("No tokens to process")
		return results
	}

	// Log at info level without showing actual tokens
	logging.Info("Processing request with %d tokens using batch size %d and %d parallel calls",
		len(tokens), p.BatchSize, p.ParallelCalls)

	// Divide tokens into batches respecting the batch size limit
	var batches [][]string
	for i := 0; i < len(tokens); i += p.BatchSize {
		end := i + p.BatchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batches = append(batches, tokens[i:end])
	}

	logging.Info("Created %d batches for processing", len(batches))

	// Process batches with limited parallelism
	semaphore := make(chan struct{}, p.ParallelCalls)
	var wg sync.WaitGroup

	for batchIndex, batch := range batches {
		wg.Add(1)

		// Anonymous function to process each batch
		go func(batchNum int, batchTokens []string) {
			defer wg.Done()

			// Acquire semaphore slot (blocking if max parallel calls reached)
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release slot when done

			logging.Info("Starting batch %d with %d tokens", batchNum, len(batchTokens))

			// Process this batch
			batchResults, err := p.ProcessTokenBatch(batchTokens, redaction)
			if err != nil {
				logging.Error("Error processing batch %d: %v", batchNum, err)
				return
			}

			// Safely merge results
			resultsMutex.Lock()
			for k, v := range batchResults {
				results[k] = v
			}
			resultsMutex.Unlock()

			logging.Info("Completed batch %d, processed %d tokens", batchNum, len(batchResults))
		}(batchIndex, batch)
	}

	// Wait for all batches to complete
	wg.Wait()
	logging.Info("All batches processed, returning %d total results", len(results))

	return results
}
