# Define the image name and tag
IMAGE_NAME := apple_log_lut:latest
CONFIG_DIR := $(shell pwd)/configs
OUTPUT_DIR := $(shell pwd)/output

# Default target: build and run
all: build run

# Build the Docker image using Buildx (multi-platform if needed)
build:
	docker buildx build --platform linux/amd64,linux/arm64 --load -t $(IMAGE_NAME) .

# Run the container, mounting config and output directories.
run:
	docker run --rm \
		-v "$(CONFIG_DIR):/app/configs" \
		-v "$(OUTPUT_DIR):/app/output" \
		$(IMAGE_NAME)

# Optionally, remove the Docker image.
clean:
	docker rmi $(IMAGE_NAME)

.PHONY: all build run clean
