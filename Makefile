.PHONY: all generate build test clean export-data

# Default target
all: generate build

# Generate protobuf Go code
generate:
	@echo "Generating protobuf Go code..."
	mkdir -p pkg/proto
	export PATH=$$PATH:$$(go env GOPATH)/bin && protoc --go_out=. --go_opt=paths=source_relative proto/names.proto
	go mod tidy

# Build the CLI tool
build: generate
	@echo "Building CLI tool..."
	go build -o bin/pii-check ./cmd/pii-check

# Export data from Python pickle to protobuf
export-data:
	@echo "Exporting data to protobuf format..."
	cd data && python3 export_to_protobuf.py

# Run tests
test:
	go test ./...

# Run benchmarks
bench:
	go test -bench=. ./...

# Clean generated files
clean:
	rm -rf pkg/proto/*.pb.go
	rm -rf bin/
	rm -rf data/*.pb.gz

# Install protobuf compiler (on macOS)
install-protoc:
	brew install protobuf

# Format code
fmt:
	go fmt ./...

# Check for issues
vet:
	go vet ./...