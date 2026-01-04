.PHONY: build run test clean dev

# Build the application
build:
	go build -o bin/assettrack ./cmd/assettrack

# Run the application
run:
	go run ./cmd/assettrack

# Run with debug logging
dev:
	go run ./cmd/assettrack -log-level debug

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/assettrack-linux ./cmd/assettrack

# Build for macOS
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/assettrack-darwin ./cmd/assettrack

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/assettrack.exe ./cmd/assettrack

# Build all platforms
build-all: build-linux build-darwin build-windows
