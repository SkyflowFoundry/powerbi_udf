#!/bin/bash

# Build the server version
echo "Building server version..."
go build -tags server -o skyflow-detokenizer-server

echo "Server binary created: skyflow-detokenizer-server"
echo "Run with ./skyflow-detokenizer-server"
