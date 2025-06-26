GOPATH := $(shell pwd)
.PHONY: clean test linux-amd64 cross-compile

all:
	@mkdir -p bin
	@cd src && go build -o ../bin/qtunnel ./qtunnel

linux-amd64:
	@mkdir -p bin
	@echo "Building for Linux AMD64..."
	@cd src && GOOS=linux GOARCH=amd64 go build -o ../bin/qtunnel-linux-amd64 ./qtunnel
	@echo "Linux AMD64 build completed: bin/qtunnel-linux-amd64"

linux-arm64:
	@mkdir -p bin
	@echo "Building for Linux ARM64..."
	@cd src && GOOS=linux GOARCH=arm64 go build -o ../bin/qtunnel-linux-arm64 ./qtunnel
	@echo "Linux ARM64 build completed: bin/qtunnel-linux-arm64"

windows-amd64:
	@mkdir -p bin
	@echo "Building for Windows AMD64..."
	@cd src && GOOS=windows GOARCH=amd64 go build -o ../bin/qtunnel-windows-amd64.exe ./qtunnel
	@echo "Windows AMD64 build completed: bin/qtunnel-windows-amd64.exe"

darwin-amd64:
	@mkdir -p bin
	@echo "Building for macOS AMD64..."
	@cd src && GOOS=darwin GOARCH=amd64 go build -o ../bin/qtunnel-darwin-amd64 ./qtunnel
	@echo "macOS AMD64 build completed: bin/qtunnel-darwin-amd64"

darwin-arm64:
	@mkdir -p bin
	@echo "Building for macOS ARM64..."
	@cd src && GOOS=darwin GOARCH=arm64 go build -o ../bin/qtunnel-darwin-arm64 ./qtunnel
	@echo "macOS ARM64 build completed: bin/qtunnel-darwin-arm64"

cross-compile: linux-amd64 linux-arm64 windows-amd64 darwin-amd64 darwin-arm64
	@echo "Cross-compilation completed for all platforms"

clean:
	@rm -fr bin pkg

test:
	@cd src && go test ./tunnel

test-quic:
	@cd src && go test ./tunnel -run TestTunnelQUIC
