BUILD_PUBLISH ?= False
BUILD_BRANCH ?= $(USER)
DOCKER_IMAGE := quay.io/signalfuse/cloudfoundry-build
DOCKER_TAG := $(DOCKER_IMAGE):$(BUILD_BRANCH)
VERSION := $(shell tr -d '\n' < VERSION)

# For Jenkins.
ifdef BASE_DIR
	ifdef JOB_NAME
		SRC_ROOT := $(BASE_DIR)/$(JOB_NAME)/cloudfoundry-integration
	endif
endif

# Fallback to PWD.
SRC_ROOT ?= $(PWD)

jar:
	./gradlew $(GRADLE_FLAGS) shadowJar

tile.yml: tile.yml.in VERSION
	m4 -DVERSION=$(VERSION) < $< > $@

tile: jar tile.yml
	tile build $(VERSION)

clean:
	rm -rf build release tile.yml

# The Docker-based build works by always attempting to build the builder image.
# If nothing has changed then this will be a quick no-op. If it's changed then
# the image will be pushed to quay so that the next builder can use the cached
# image. The advantage of this method as opposed to having a separate image
# build step is that the cached image includes all of the application
# dependencies so that the actual build step can be done completely offline
# which makes it fast.
docker-build:
	docker run -t --rm -v $(SRC_ROOT):/opt/src $(DOCKER_TAG) \
		sh -c "cd /opt/src && make tile GRADLE_FLAGS=\"--offline --no-daemon --console plain\""

docker-deps:
# Build the build image with dependencies baked in.
	rm -rf build/docker-deps
	mkdir -p build/docker-deps
# Copy Docker skeleton.
	cp -r docker/* build/docker-deps
# Copy build files to source load dependencies from.
	cp -r build.gradle settings.gradle gradlew gradle VERSION build/docker-deps
	docker build -t $(DOCKER_TAG) build/docker-deps
ifeq ($(BUILD_PUBLISH), True)
	docker push $(DOCKER_TAG)
endif

jenkins-build:
# Pull latest before doing build so that if it's already been
# built we don't try to rebuild it in make-deps. Image may not
# yet exist so ignore errors. Would be better to actually check
# explicitly for existence as this could cover up things like
# network errors.
	docker pull $(DOCKER_TAG) || true

	make docker-deps
	make docker-build

ifeq ($(BUILD_PUBLISH), True)
	aws s3 cp product/*.pivotal s3://signalfx-cloudfoundry/builds/ \
		--cache-control="max-age=0, no-cache" \
		--acl private
endif

build-and-push: tile
	pcf import product/signalfx-agent-$(VERSION).pivotal
	pcf install signalfx-agent $(VERSION)

.PHONY: jar tile clean build-and-push docker-deps docker-build docker-jar jenkins-build
