#!/bin/bash

# qTunnel QUIC Demo Script
# This script demonstrates how to use qTunnel with QUIC protocol
# Note: QUIC mode uses TCP for frontend listening and QUIC for backend transport

echo "=== qTunnel QUIC Demo ==="
echo "This demo shows how to set up a secure tunnel using QUIC protocol"
echo "Architecture: TCP frontend -> QUIC backend transport"
echo ""

# Check if qtunnel binary exists
if [ ! -f "../bin/qtunnel" ]; then
    echo "Error: qtunnel binary not found. Please run 'make' first."
    exit 1
fi

echo "Step 1: Setting up a simple HTTP server on port 8080..."
# Start a simple HTTP server in background
python3 -m http.server 8080 &
HTTP_PID=$!
echo "HTTP server started with PID: $HTTP_PID"

echo ""
echo "Step 2: Starting qTunnel server (QUIC mode) on port 9001..."
echo "Note: Server listens on TCP but uses QUIC for backend connections"
echo "Command: ../bin/qtunnel -listen=:9001 -backend=127.0.0.1:8080 -secret=demo123 -crypto=aes256cfb -quic=true"
# Start qtunnel server in background
../bin/qtunnel -listen=:9001 -backend=127.0.0.1:8080 -secret=demo123 -crypto=aes256cfb -quic=true &
SERVER_PID=$!
echo "qTunnel server started with PID: $SERVER_PID"

echo ""
echo "Step 3: Starting qTunnel client (QUIC mode) on port 9002..."
echo "Note: Client listens on TCP but uses QUIC to connect to server"
echo "Command: ../bin/qtunnel -listen=:9002 -backend=127.0.0.1:9001 -clientmode=true -secret=demo123 -crypto=aes256cfb -quic=true"
# Start qtunnel client in background
../bin/qtunnel -listen=:9002 -backend=127.0.0.1:9001 -clientmode=true -secret=demo123 -crypto=aes256cfb -quic=true &
CLIENT_PID=$!
echo "qTunnel client started with PID: $CLIENT_PID"

echo ""
echo "Step 4: Waiting for services to start..."
sleep 3

echo ""
echo "Step 5: Testing the tunnel..."
echo "Direct access to HTTP server: curl -s http://127.0.0.1:8080 | head -1"
curl -s http://127.0.0.1:8080 | head -1

echo ""
echo "Access through qTunnel (encrypted): curl -s http://127.0.0.1:9002 | head -1"
curl -s http://127.0.0.1:9002 | head -1

echo ""
echo "Demo completed! The tunnel is working."
echo "Traffic flow: curl -> qTunnel client (TCP:9002) -[QUIC]-> qTunnel server (TCP:9001) -> HTTP server (port 8080)"
echo "QUIC is used for the encrypted tunnel between client and server"
echo ""
echo "Press Enter to stop all services..."
read

echo "Stopping services..."
kill $HTTP_PID $SERVER_PID $CLIENT_PID 2>/dev/null
echo "All services stopped."
echo "Demo finished!"