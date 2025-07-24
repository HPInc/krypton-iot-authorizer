all: build

build: build-binaries

# Build the binaries for the lambda.
build-binaries:
	$(GOBUILD) -ldflags $(LDFLAGS) \
	-o $(BINARY_DIR)/$(BINARY_NAME)

# Create a docker image for the lambda.
docker-image:
	docker build -t $(DOCKER_IMAGE) -f Dockerfile .

include common.mk
.PHONY: docker-image tag push check_changes clean
