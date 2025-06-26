#!/bin/bash

# qTunnel TCP Demo Script
# This script demonstrates how to use qTunnel with traditional TCP protocol

echo "=== qTunnel TCP Demo ==="
echo "This demo shows how to set up a secure tunnel using TCP protocol"
echo ""

# Check if qtunnel binary exists
if [ ! -f "../bin/qtunnel" ]; then
    echo "Error: qtunnel binary not found. Please run 'make' first."
    exit 1
fi

echo "Step 1: Setting up a simple HTTP server on port 8081..."
# Start a simple HTTP server in background
python3 -m http.server 8081 &
HTTP_PID=$!
echo "HTTP server started with PID: $HTTP_PID"

echo ""
echo "Step 2: Starting qTunnel server (TCP mode) on port 9003..."
echo "Command: ../bin/qtunnel -listen=:9003 -backend=127.0.0.1:8081 -secret=demo123 -crypto=aes256cfb"
# Start qtunnel server in background
../bin/qtunnel -listen=:9003 -backend=127.0.0.1:8081 -secret=demo123 -crypto=aes256cfb &
SERVER_PID=$!
echo "qTunnel server started with PID: $SERVER_PID"

echo ""
echo "Step 3: Starting qTunnel client (TCP mode) on port 9004..."
echo "Command: ../bin/qtunnel -listen=:9004 -backend=127.0.0.1:9003 -clientmode=true -secret=demo123 -crypto=aes256cfb"
# Start qtunnel client in background
../bin/qtunnel -listen=:9004 -backend=127.0.0.1:9003 -clientmode=true -secret=demo123 -crypto=aes256cfb &
CLIENT_PID=$!
echo "qTunnel client started with PID: $CLIENT_PID"

echo ""
echo "Step 4: Waiting for services to start..."
sleep 3

echo ""
echo "Step 5: Testing the tunnel..."
echo "Direct access to HTTP server: curl -s http://127.0.0.1:8081 | head -1"
curl -s http://127.0.0.1:8081 | head -1

echo ""
echo "Access through qTunnel (encrypted): curl -s http://127.0.0.1:9004 | head -1"
curl -s http://127.0.0.1:9004 | head -1

echo ""
echo "Demo completed! The TCP tunnel is working."
echo "Traffic flow: curl -> qTunnel client (port 9004) -> qTunnel server (port 9003) -> HTTP server (port 8081)"
echo ""
echo "Press Enter to stop all services..."
read

echo "Stopping services..."
kill $HTTP_PID $SERVER_PID $CLIENT_PID 2>/dev/null
echo "All services stopped."
echo "Demo finished!"