#!/bin/bash

# Define the server command and the endpoint
SERVER_CMD="go run ./cmd/... -port 9000"
HEALTH_ENDPOINT="http://localhost:9000/"
WEATHER_ENDPOINT="http://localhost:9000/weather/35.7596,-79.0193"
MAX_ATTEMPTS=10
WAIT_SECONDS=2

# 1. Start the server in the background
echo "Starting server..."
# The '&' runs the command in a background process
# '>/dev/null 2>&1' redirects stdout and stderr to prevent clutter in the main script output
$SERVER_CMD >/dev/null 2>&1 &
# Capture the Process ID (PID) of the last background command
SERVER_PID=$!
echo "Server started with PID $SERVER_PID"

# 2. Wait for the server to be ready
echo "Waiting for server to become available..."
attempts=0
while [ $attempts -lt $MAX_ATTEMPTS ]; do
    # Use curl to check the server status.
    # -s: silent, -o /dev/null: discard output, -f: fail silently on HTTP errors
    if curl -s -f -o /dev/null "$HEALTH_ENDPOINT"; then
        echo "Server is up and running!"
        break
    fi
    echo "Attempt $((attempts + 1))/$MAX_ATTEMPTS: Server unavailable, sleeping for $WAIT_SECONDS seconds..."
    sleep $WAIT_SECONDS
    ((attempts++))
done

# Check if the server ever started
if [ $attempts -eq $MAX_ATTEMPTS ]; then
    echo "Error: Server failed to start within the time limit."
    kill "$SERVER_PID"
    exit 1
fi

# 3. Send the curl request
echo "Sending curl request to $WEATHER_ENDPOINT"
curl "$WEATHER_ENDPOINT"
echo "" # Add a newline for cleaner output

# 4. Clean up: Kill the background server process
echo "Shutting down server (PID $SERVER_PID)..."
kill "$SERVER_PID"
echo "Server stopped."