# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

# Default number of clusters to run
N ?= 3
PN ?= 2

# Use .PHONY to declare targets that are not files.
.PHONY: all build test run runp stop clean gen

# Build all etcds and run tests by default
all: test build

# Build etcd directories and binaries
build:
	@echo "=> Building etcd binary..."
	@mkdir -p bin/etcd/
	@rm -f bin/etcd/etcd # Use -f to avoid errors if the file doesn't exist
	@$(GOBUILD) -o bin/etcd/etcd ./cmd/etcd/main.go

# Run all etcds
run:
	@echo "=> Running $(N) etcd instance(s)..."
	@for i in $$(shell seq 1 $(N)); do \
	   echo "   -> Starting etcd$$i..."; \
	   (cd bin/etcd/ && ./etcd -c ../../cluster_config/node$$i/config.yml) & \
	done

# Run etcds for performance testing
runp:
	@echo "=> Running $(PN) etcd instance(s) for performance tests..."
	@for i in $$(shell seq 1 $(PN)); do \
	   echo "   -> Starting etcd$$i..."; \
	   (cd bin/etcd/ && ./etcd -c ../../pcluster_config/node$$i/config.yml) & \
	done

# Stop all running etcd instances
stop:
	@echo "=> Stopping all etcd instances..."
	@-pkill -f ./etcd # The '-' before pkill ignores errors if no process is found

# Clean up build artifacts
clean:
	@echo "=> Cleaning up build artifacts..."
	@rm -rf bin/etcd

# Generate protobuf code
gen:
	@echo "=> Generating protobuf code..."
	@protoc --go_out=.. --go-grpc_out=.. --go-grpc_opt=require_unimplemented_servers=false -I. -Iproto proto/pb/pb.proto

# Placeholder for tests - you can add your test commands here
test:
	@echo "=> Running tests..."
	@# Example: $(GOTEST) ./...