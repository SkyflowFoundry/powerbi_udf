#!/bin/bash

# Build the Lambda function
echo "Building Lambda function..."
GOOS=linux GOARCH=amd64 go build -tags lambda -o bootstrap lambda.go

# Create deployment package
echo "Creating deployment package..."
zip skyflow-detokenizer-lambda.zip bootstrap
zip -r skyflow-detokenizer-lambda.zip config/

echo "Lambda deployment package created: skyflow-detokenizer-lambda.zip"
echo "Upload this file to AWS Lambda and set the handler to 'bootstrap'"
