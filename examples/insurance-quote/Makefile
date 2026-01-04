.PHONY: build run test clean dev

# Build the application
build:
	go build -o bin/insurequote ./cmd/insurequote

# Run the application
run:
	go run ./cmd/insurequote

# Run with a specific port
run-port:
	PORT=$(PORT) go run ./cmd/insurequote

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/insurequote-linux ./cmd/insurequote

# Build for macOS
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/insurequote-darwin ./cmd/insurequote

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/insurequote.exe ./cmd/insurequote

# Build all platforms
build-all: build-linux build-darwin build-windows
