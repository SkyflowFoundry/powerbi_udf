// handler/handler.go

package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"skyflow_detokenizer/internal/logging"
	"skyflow_detokenizer/internal/models"
	"skyflow_detokenizer/internal/processor"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DetokenizeHandler creates a new HTTP handler for detokenization requests
func DetokenizeHandler(p *processor.TokenProcessor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log the incoming request at info level (without sensitive data)
		logging.Info("Received detokenization request from %s", c.ClientIP())
		
		// Only log headers at debug level as they might contain sensitive data
		logging.Debug("Received headers: %v", c.Request.Header)

		// Read the raw request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logging.Error("Error reading body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read request body"})
			return
		}

		// Log the raw request only at debug level
		logging.Debug("Received raw request: %s", string(body))

		// Restore the request body since we've read it once
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Parse the JSON request into our request structure
		var req models.DetokenizeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logging.Error("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format: " + err.Error(),
			})
			return
		}

		// Extract tokens from token objects
		tokens := make([]string, len(req.Tokens))
		for i, t := range req.Tokens {
			tokens[i] = t.Token
		}

		// Log token count at info level (without exposing actual tokens)
		logging.Info("Processing request with %d tokens and redaction type: %s",
			len(tokens), req.Redaction)

		// Process the tokens
		results := p.ProcessTokens(tokens, req.Redaction)

		// Log result count at info level
		logging.Info("Request processed successfully, returning %d results", len(results))
		
		// Log the detailed results only at debug level
		logging.Debug("Processed results: %v", results)

		// Convert results to JSON before sending
		responseJSON, err := json.Marshal(results)
		if err != nil {
			logging.Error("Error marshaling response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response"})
			return
		}

		// Set explicit headers for better SQL Server compatibility
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Header("Content-Length", strconv.Itoa(len(responseJSON)))
		c.Header("Connection", "close") // Ensure connection is closed after response

		// Write the response directly instead of using c.JSON
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(responseJSON)
		
		// Log completion at info level
		logging.Info("Response sent successfully")
	}
}
