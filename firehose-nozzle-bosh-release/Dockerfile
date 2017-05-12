FROM cfplatformeng/tile-generator

ENTRYPOINT ["/bin/cat"]

RUN apk update && \
    apk add go=1.7.3-r0 wget file &&\
	rm /var/cache/apk/*

RUN gem install bosh_cli --no-ri --no-rdoc

RUN cd /tmp &&\
    wget https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz &&\
	tar -xf glide*.tar.gz &&\
	mv linux-amd64/glide /usr/bin/glide &&\
	wget https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz

ENV GOPATH=/tmp/go

# symlink our code into the GOPATH
RUN mkdir -p $GOPATH/src/github.com/signalfx/ &&\
	ln -s /opt/src $GOPATH/src/github.com/signalfx/cloudfoundry-integration

ADD main.go glide.yaml glide.lock /opt/src/
ADD metrics /opt/src/metrics
# Glide wants the test deps too
ADD testhelpers /opt/src/testhelpers

WORKDIR /opt/src
RUN cd $GOPATH/src/github.com/signalfx/cloudfoundry-integration &&\
    glide get --non-interactive github.com/signalfx/cloudfoundry-integration

ADD bosh-release /opt/src/bosh-release

# BOSH release bundler doesn't follow symlinks so must make a copy
RUN cp -r main.go metrics vendor \
	    bosh-release/src/signalfx_agent/src/github.com/signalfx/cloudfoundry-integration/

ADD resources /opt/src/resources
ADD tile.yml tile-history.yml VERSION /opt/src/
ADD bin /opt/src/bin

RUN cd bosh-release &&\
    bosh add blob /tmp/go1.8.1.linux-amd64.tar.gz golang1.8 &&\
    bosh create release --force --with-tarball | tee /tmp/release.log

RUN tile build $(cat VERSION)

