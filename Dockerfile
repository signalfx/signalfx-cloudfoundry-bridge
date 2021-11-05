FROM golang:1.17 AS builder

WORKDIR /tmp/signalfx-cloudfoundry-bridge

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./
COPY metrics ./metrics
COPY testhelpers ./testhelpers
COPY Makefile ./

RUN make signalfx-bridge

FROM busybox:1.34

ENTRYPOINT /bin/signalfx-bridge
COPY --from=builder /tmp/signalfx-cloudfoundry-bridge/signalfx-bridge /bin/signalfx-bridge
