FROM ghcr.io/hpinc/krypton/krypton-go-builder AS builder

ADD . /go/src/iot_authorizer
WORKDIR /go/src/iot_authorizer

# build the source
RUN make gosec build-binaries

# use a minimal alpine image for services
FROM ghcr.io/hpinc/krypton/krypton-go-base

# set working directory
WORKDIR /go/bin

COPY --from=builder /go/src/iot_authorizer/bin .

USER 1001

# run the binary
CMD ["/go/bin/aws-iot-authorizer"]
