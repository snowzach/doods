ARG BUILDER_TAG="latest"
FROM snowzach/doods-base:$BUILDER_TAG as base

# Install Go
ENV GO_VERSION "1.13.3"
RUN curl -kLo go${GO_VERSION}.linux-${GO_ARCH}.tar.gz https://dl.google.com/go/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-${GO_ARCH}.tar.gz && \
    rm go${GO_VERSION}.linux-${GO_ARCH}.tar.gz

FROM debian:buster-slim as build

RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config zip zlib1g-dev unzip wget bash-completion git curl \
    build-essential patch g++ python python-future python3 ca-certificates \
    libc6-dev libstdc++6 libusb-1.0-0

# Copy all libraries, includes and go
COPY --from=base /usr/local/. /usr/local/.
COPY --from=base /opt/tensorflow /opt/tensorflow 

ENV GOOS=linux
ENV CGO_ENABLED=1
ENV CGO_CFLAGS=-I/opt/tensorflow
ENV PATH /usr/local/go/bin:/go/bin:${PATH}
ENV GOPATH /go

# Create the build directory
RUN mkdir /build
WORKDIR /build
