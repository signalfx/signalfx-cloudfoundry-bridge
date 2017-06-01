DOCKER_BUILD_IMAGE := cf-bridge-builder

all: metrics main.go glide.yaml glide.lock Dockerfile testhelpers
	docker build -t $(DOCKER_BUILD_IMAGE) .
	docker run --rm $(DOCKER_BUILD_IMAGE) > signalfx-bridge-linux-amd64
	chmod +x signalfx-bridge-linux-amd64

