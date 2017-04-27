BUILD_PUBLISH ?= False
BUILD_BRANCH ?= $(USER)
DOCKER_IMAGE := quay.io/signalfuse/cloudfoundry-integration
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

tile.yml: tile.yml.in VERSION
	m4 -DVERSION=$(VERSION) < $< > $@

tile: jar tile.yml
	tile build $(VERSION)

clean:
	rm -rf build release tile.yml

build-image:
	docker build -t $(DOCKER_TAG) .

docker-build:
	docker run -t --rm -v $(SRC_ROOT):/opt/src $(DOCKER_TAG) \
		sh -c "cd /opt/src && make tile GRADLE_FLAGS=\"--offline --no-daemon --console plain\""

docker-deps:
# Copy build files to source load dependencies from.
	docker build -t $(DOCKER_TAG) .
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
