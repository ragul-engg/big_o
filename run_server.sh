#!/bin/bash

# Check if a port number is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <port>"
  exit 1
fi

PORT=$1

# Set the CURRENT_NODE_IP dynamically based on the provided port
export CURRENT_NODE_IP="http://localhost:$PORT"

# Load server IPs (if needed)
source ./serverIps.sh

# Run the Go application with the provided port
go build -o big_o .

./big_o --port $PORT >>"logs_${PORT}.log"
