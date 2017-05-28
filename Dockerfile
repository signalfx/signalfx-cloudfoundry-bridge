FROM ubuntu:yakkety

ENV GOPATH=/go PATH=$PATH:/go/bin

RUN apt-get update &&\
    apt-get install -yq golang curl git &&\
	mkdir -p $GOPATH/bin $GOPATH/src/github.com/signalfx/cloudfoundry-bridge &&\
	curl https://glide.sh/get | sh

WORKDIR $GOPATH/src/github.com/signalfx/cloudfoundry-bridge

ADD glide* main.go ./
ADD metrics ./metrics
ADD testhelpers ./testhelpers

RUN glide install github.com/signalfx/cloudfoundry-bridge

RUN go install

CMD cat $GOPATH/bin/cloudfoundry-bridge
