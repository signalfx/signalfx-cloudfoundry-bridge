EXECUTABLE_NAME := signalfx-bridge
DOCKER_IMAGE := signalfx-bridge

.PHONY: signalfx-bridge
signalfx-bridge:
	CGO_ENABLED=0 go build -o $(EXECUTABLE_NAME) .

.PHONY: signalfx-bridge-linux
signalfx-bridge-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(EXECUTABLE_NAME)-linux .

.PHONY: docker
docker:
	docker build -t $(DOCKER_IMAGE) .

