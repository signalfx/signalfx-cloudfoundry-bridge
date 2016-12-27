jar:
	./gradlew shadowJar

tile: jar
	tile build

build-and-push: tile
	version=`ls product/*.pivotal | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+'` && \
	pcf import product/*.pivotal && \
	pcf install signalfx-agent $$version

.PHONY: jar tile build-and-push
