# Skyflow Detokenizer Service

A secure and efficient service for detokenizing Skyflow tokens with configurable logging and error handling capabilities.

## Overview

The Skyflow Detokenizer Service is designed to securely process token detokenization requests through the Skyflow API. It supports batch processing, parallel API calls, and has configurable logging levels to ensure sensitive data is properly handled.

Key features:
- Batch processing of tokens with configurable batch size
- Parallel API calls for improved performance
- Configurable logging (info/debug levels and console/file output)
- Continues processing on error with the `ContinueOnError` flag
- Multiple deployment options (local, Docker, AWS Lambda)

## Configuration

Configuration is managed through a YAML file located at `config/config.yaml`:

```yaml
skyflow:
  vault_url: "https://your-vault-url.vault.skyflowapis.com/v1/vaults/your-vault-id"
  bearer_token: "your-bearer-token"
  batch_size: 100
  parallel_calls: 10
  request_timeout_seconds: 30

server:
  port: 8080

logging:
  level: "info"        # "info" or "debug"
  log_to_file: false   # true to log to file, false for console
  file_path: "./logs/skyflow_detokenizer.log"
```

### Configuration Options

#### Skyflow Configuration
- `vault_url`: Your Skyflow vault URL
- `bearer_token`: Authentication token for Skyflow API
- `batch_size`: Maximum number of tokens to process in one API call
- `parallel_calls`: Maximum number of parallel API calls
- `request_timeout_seconds`: Timeout in seconds for API calls

#### Server Configuration
- `port`: The port on which the HTTP server will listen

#### Logging Configuration
- `level`: Log level ("info" or "debug")
  - `info`: Only shows operational information (API calls, success/failure)
  - `debug`: Shows detailed information including tokens and responses
- `log_to_file`: Whether to log to a file (true) or console (false)
- `file_path`: Path to the log file when logging to a file

## API Usage

The service exposes a single endpoint for detokenization:

### POST /detokenize

Request body:
```json
{
  "tokens": [
    {"token": "token1"},
    {"token": "token2"},
    {"token": "token3"}
  ],
  "redaction": "PLAIN_TEXT"
}
```

Response:
```json
{
  "token1": "value1",
  "token2": "value2",
  "token3": "Token Not Found"
}
```

## Deployment Options

### Local Deployment

1. Ensure Go is installed on your system (version 1.16 or higher)
2. Update the `config/config.yaml` file with your Skyflow credentials
3. Build and run the application:

```bash
go mod init skyflow_detokenizer
go mod tidy
go build
./skyflow_detokenizer
```

### Docker Deployment

1. Update the `config/config.yaml` file with your Skyflow credentials
2. Build and run the Docker container:

```bash
# Build the Docker image
docker build -t skyflow_detokenizer .

# Run the container
docker run -p 8080:8080 -v $(pwd)/config:/app/config skyflow_detokenizer
```

Alternatively, use Docker Compose:

```bash
docker-compose up -d
```

### AWS Lambda Deployment

1. Update the `config/config.yaml` file with your Skyflow credentials
2. Install the AWS CLI and AWS SAM CLI
3. Build and deploy the Lambda function:

```bash
# Build the Lambda function
GOOS=linux GOARCH=amd64 go build -o bootstrap lambda.go

# Deploy using SAM CLI
sam deploy --guided
```

During the guided deployment, you'll be prompted to provide:
- Stack name
- AWS Region
- Confirmation of IAM role creation

#### Manual Deployment Steps

If you prefer to deploy manually:

1. Build the Lambda function:
```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap lambda.go
```

2. Create a deployment package:
```bash
zip skyflow-detokenizer-lambda.zip bootstrap
zip -r skyflow-detokenizer-lambda.zip config/
```

3. Create or update the Lambda function using AWS CLI:
```bash
# Create a new Lambda function
aws lambda create-function \
  --function-name skyflow-detokenizer \
  --runtime provided.al2 \
  --handler bootstrap \
  --zip-file fileb://skyflow-detokenizer-lambda.zip \
  --role arn:aws:iam::<account-id>:role/<lambda-execution-role>

# Or update an existing function
aws lambda update-function-code \
  --function-name skyflow-detokenizer \
  --zip-file fileb://skyflow-detokenizer-lambda.zip
```

4. Create an API Gateway trigger:
```bash
aws lambda add-permission \
  --function-name skyflow-detokenizer \
  --statement-id apigateway \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com
```

## Security Considerations

- The service is designed to protect sensitive data by only logging it at the debug level
- Bearer tokens and other credentials should be kept secure
- Consider using AWS Secrets Manager or similar services for managing credentials in production
- Enable HTTPS when deploying to production environments

## Troubleshooting

If you encounter issues:

1. Check the logs (console or log file depending on your configuration)
2. Verify your Skyflow credentials in the config file
3. Ensure your network allows connections to the Skyflow API
4. For Lambda deployments, check the CloudWatch logs

