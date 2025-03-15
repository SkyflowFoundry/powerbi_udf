FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o skyflow-detokenizer main.go

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/skyflow-detokenizer .
COPY --from=builder /app/config ./config

# Create a directory for logs if needed
RUN mkdir -p /app/logs

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./skyflow-detokenizer"]
