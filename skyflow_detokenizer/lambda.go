//go:build lambda
// +build lambda

package main

import (
	"context"
	"encoding/json"
	"skyflow_detokenizer/internal/config"
	"skyflow_detokenizer/internal/logging"
	"skyflow_detokenizer/internal/models"
	"skyflow_detokenizer/internal/processor"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// LambdaHandler handles AWS Lambda requests
func LambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to load configuration"}`,
		}, nil
	}

	// Initialize logging
	if err := logging.InitLogger(cfg.Logging.Level, cfg.Logging.LogToFile, cfg.Logging.FilePath); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to initialize logging"}`,
		}, nil
	}
	defer logging.Close()

	// Log request received (without sensitive data)
	logging.Info("Received Lambda request")
	logging.Debug("Request body: %s", request.Body)

	// Parse the request body
	var req models.DetokenizeRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		logging.Error("Failed to parse request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Invalid request format"}`,
		}, nil
	}

	// Extract tokens from token objects
	tokens := make([]string, len(req.Tokens))
	for i, t := range req.Tokens {
		tokens[i] = t.Token
	}

	// Log token count (without exposing actual tokens)
	logging.Info("Processing request with %d tokens", len(tokens))

	// Initialize the token processor
	p := processor.NewTokenProcessor(
		cfg.Skyflow.VaultURL,
		cfg.Skyflow.BearerToken,
		cfg.Skyflow.BatchSize,
		cfg.Skyflow.ParallelCalls,
		cfg.Skyflow.RequestTimeoutSecs,
	)

	// Process the tokens
	results := p.ProcessTokens(tokens, req.Redaction)

	// Log result count
	logging.Info("Request processed successfully, returning %d results", len(results))
	logging.Debug("Processed results: %v", results)

	// Convert results to JSON
	responseJSON, err := json.Marshal(results)
	if err != nil {
		logging.Error("Error marshaling response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to generate response"}`,
		}, nil
	}

	// Return the response
	logging.Info("Response sent successfully")
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJSON),
	}, nil
}

func main() {
	lambda.Start(LambdaHandler)
}
