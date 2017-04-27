FROM golang:1.8

ADD . /go/src/github.com/signalfx/cloudfoundry-integration
RUN go get github.com/signalfx/cloudfoundry-integration

ENTRYPOINT /go/bin/cloudfoundry-integration

# OpenTSDB port for BOSH HM metrics
EXPOSE 13321
