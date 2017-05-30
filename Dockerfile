FROM ubuntu:yakkety

ENV GOPATH=/go PATH=$PATH:/usr/local/go/bin:/go/bin

RUN apt-get update &&\
    apt-get install -yq wget curl git &&\
	wget -O /tmp/go.tar.gz https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz &&\
	tar -C /usr/local -xzf /tmp/go.tar.gz &&\
	mkdir -p $GOPATH/bin $GOPATH/src/github.com/signalfx/signalfx-bridge &&\
	curl https://glide.sh/get | sh

WORKDIR $GOPATH/src/github.com/signalfx/signalfx-bridge

COPY glide* main.go ./
COPY metrics ./metrics
COPY testhelpers ./testhelpers

RUN glide install github.com/signalfx/signalfx-bridge

RUN go install

CMD cat $GOPATH/bin/signalfx-bridge
