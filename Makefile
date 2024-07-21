# Variables
BUILD_FOLDER=build
PROTO_PATH=internal/proto
DOCKERFILE_CONTROLLER=docker/Dockerfile.controller
DOCKERFILE_AGENT=docker/Dockerfile.agent
DOCKERFILE_POC_MODULE=poc-module/Dockerfile
IMAGE_CONTROLLER=dmapz-agent-controller
IMAGE_AGENT=dmapz-agent
IMAGE_POC_MODULE=poc-module
EXPORT_CONTROLLER=build/image-agent-controller.tar
EXPORT_AGENT=build/image-agent.tar
EXPORT_POC_MODULE=build/image-poc-module.tar

# Default target
all: build

# Build Docker image and Go binary
build:
	@echo "Building Docker images and preparing deployment files..."
	mkdir -p $(BUILD_FOLDER)
	docker build -f $(DOCKERFILE_CONTROLLER) -t $(IMAGE_CONTROLLER) .
	docker build -f $(DOCKERFILE_AGENT) -t $(IMAGE_AGENT) .
	docker build -f $(DOCKERFILE_POC_MODULE) -t $(IMAGE_POC_MODULE) .
	docker save $(IMAGE_CONTROLLER):latest | gzip > $(EXPORT_CONTROLLER)
	docker save $(IMAGE_AGENT):latest | gzip > $(EXPORT_AGENT)
	docker save $(IMAGE_POC_MODULE):latest | gzip > $(EXPORT_POC_MODULE)

# Clean the Go binary, Docker image, and call cleanup script
clean:
	@echo "Cleaning up..."
	rm -r $(BUILD_FOLDER)
	./scripts/cleanup.sh
	docker rmi $(IMAGE_POC_MODULE) $(IMAGE_CONTROLLER) $(IMAGE_AGENT) || true

# Run all Go tests in the project
test:
	@echo "Running Go tests..."
	go test ./...

# Subtargets for the "run" command
run: $(TARGET)

# Run controller by calling deploy script and passing arguments as environment variables
controller:
	@echo "Deploying main node inftastructure and starting agent controller..."
	@$(env) ./scripts/main_node_start.sh

# Run agent by calling deploy script and setting NODE_NUM environment variable
agent:
	@echo "Deploying node inftastructure and starting agent..."
	@$(env) ./scripts/node_start.sh

# Generate protofiles
generate:
	@echo "Generating protofiles..."
	protoc --proto_path=$(PROTO_PATH) --go_out=$(PROTO_PATH)/. --go_opt=paths=source_relative --go-grpc_out=$(PROTO_PATH)/. --go-grpc_opt=paths=source_relative $(PROTO_PATH)/*.proto

# Phony targets to avoid conflicts with files named the same
.PHONY: all build clean test run controller node agent generate
