#!/bin/bash

# Test script for QUIC connection fix
echo "=== Testing QUIC Connection Fix ==="

# Check if qtunnel binary exists
if [ ! -f "./bin/qtunnel" ]; then
    echo "Error: qtunnel binary not found. Please run 'make' first."
    exit 1
fi

# Kill any existing processes on our ports
echo "Cleaning up any existing processes..."
lsof -ti:8085,9005,9006 | xargs kill -9 2>/dev/null || true
sleep 1

# Start HTTP server
echo "Starting HTTP server on port 8085..."
python3 -m http.server 8085 &
HTTP_PID=$!
sleep 2

# Start QUIC server
echo "Starting QUIC server on port 9005..."
./bin/qtunnel -listen=:9005 -backend=127.0.0.1:8085 -secret=mysecret -quic=true &
SERVER_PID=$!
sleep 2

# Start QUIC client
echo "Starting QUIC client on port 9006..."
./bin/qtunnel -listen=:9006 -backend=127.0.0.1:9005 -clientmode=true -secret=mysecret -quic=true &
CLIENT_PID=$!
sleep 3

# Test the connection
echo "Testing QUIC tunnel..."
RESULT=$(curl -s --connect-timeout 5 http://127.0.0.1:9006 | head -1)

if [[ $RESULT == *"DOCTYPE"* ]]; then
    echo "✅ QUIC tunnel working successfully!"
    echo "Response: $RESULT"
else
    echo "❌ QUIC tunnel failed"
    echo "Response: $RESULT"
fi

# Cleanup
echo "Cleaning up..."
kill $HTTP_PID $SERVER_PID $CLIENT_PID 2>/dev/null
wait 2>/dev/null

echo "Test completed."