# qTunnel Examples

This directory contains example scripts demonstrating how to use qTunnel with both TCP and QUIC protocols.

## Prerequisites

1. Build qTunnel first:
   ```bash
   cd ..
   make
   ```

2. Ensure you have Python 3 installed (for the demo HTTP server)

## Available Examples

### TCP Demo (`tcp_demo.sh`)

Demonstrates the traditional TCP-based tunnel:

```bash
./tcp_demo.sh
```

This script:
- Starts an HTTP server on port 8081
- Creates a qTunnel server on port 9003 (connects to HTTP server)
- Creates a qTunnel client on port 9004 (connects to qTunnel server)
- Tests the encrypted tunnel by making HTTP requests

### QUIC Demo (`quic_demo.sh`)

Demonstrates the QUIC-based tunnel with hybrid architecture:

```bash
./quic_demo.sh
```

This script:
- Starts an HTTP server on port 8080
- Creates a qTunnel server on port 9001 with QUIC enabled (TCP frontend, QUIC backend)
- Creates a qTunnel client on port 9002 with QUIC enabled (TCP frontend, QUIC backend)
- Tests the encrypted tunnel by making HTTP requests

**QUIC Architecture**: Both client and server listen on TCP for compatibility, but use QUIC protocol for the encrypted tunnel communication between them.

## Traffic Flow

### TCP Demo
```
Client Request -> qTunnel Client (TCP) -[TCP+Encryption]-> qTunnel Server (TCP) -> Backend Service
```

### QUIC Demo
```
Client Request -> qTunnel Client (TCP frontend) -[QUIC+Encryption]-> qTunnel Server (TCP frontend) -> Backend Service
```

**Key Difference**: QUIC mode uses QUIC protocol for the tunnel communication between client and server, providing better performance, connection migration, and built-in security features, while maintaining TCP compatibility for client connections.

## Notes

- The QUIC implementation requires proper TLS certificate setup for production use
- Both demos use AES256-CFB encryption with the password "demo123"
- Press Enter in the demo scripts to stop all services
- Each demo uses different ports to avoid conflicts

## Manual Testing

You can also test qTunnel manually:

### TCP Mode
```bash
# Terminal 1: Start backend service
python3 -m http.server 8000

# Terminal 2: Start qTunnel server
../bin/qtunnel -listen=:9001 -backend=127.0.0.1:8000 -secret=mysecret -crypto=rc4

# Terminal 3: Start qTunnel client
../bin/qtunnel -listen=:9002 -backend=127.0.0.1:9001 -clientmode=true -secret=mysecret -crypto=rc4

# Terminal 4: Test the tunnel
curl http://127.0.0.1:9002
```

### QUIC Mode
```bash
# Terminal 1: Start backend service
python3 -m http.server 8000

# Terminal 2: Start qTunnel server with QUIC
../bin/qtunnel -listen=:9001 -backend=127.0.0.1:8000 -secret=mysecret -crypto=rc4 -quic=true

# Terminal 3: Start qTunnel client with QUIC
../bin/qtunnel -listen=:9002 -backend=127.0.0.1:9001 -clientmode=true -secret=mysecret -crypto=rc4 -quic=true

# Terminal 4: Test the tunnel
curl http://127.0.0.1:9002
```

Note: In QUIC mode, both client and server still listen on TCP ports for client connections, but use QUIC protocol for the encrypted tunnel between them.