EXECUTABLE := doods
GITVERSION := $(shell git describe --dirty --always --tags --long)
GOPATH ?= ${HOME}/go
TAG ?= latest
PACKAGENAME := $(shell go list -m -f '{{.Path}}')
TOOLS := ${GOPATH}/src/github.com/gogo/protobuf/proto \
	${GOPATH}/bin/protoc-gen-gogoslick \
	${GOPATH}/bin/protoc-gen-grpc-gateway \
	${GOPATH}/bin/protoc-gen-swagger
export PROTOBUF_INCLUDES = -I. -I/usr/include -I${GOPATH}/src -I$(shell go list -e -f '{{.Dir}}' .) -I$(shell go list -e -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/runtime)/../third_party/googleapis
PROTOS := ./server/rpc/version.pb.gw.go \
	./odrpc/rpc.pb.gw.go

.PHONY: default
default: ${EXECUTABLE}

# This is all the tools required to compile, test and handle protobufs
tools: ${TOOLS}

${GOPATH}/src/github.com/gogo/protobuf/proto:
	GO111MODULE=off go get github.com/gogo/protobuf/proto

${GOPATH}/bin/protoc-gen-gogoslick:
	go get github.com/gogo/protobuf/protoc-gen-gogoslick

${GOPATH}/bin/protoc-gen-grpc-gateway:
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

${GOPATH}/bin/protoc-gen-swagger:
	go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

# Handle all grpc endpoint protobufs
%.pb.gw.go: %.proto
	protoc ${PROTOBUF_INCLUDES} --gogoslick_out=paths=source_relative,plugins=grpc:. --grpc-gateway_out=paths=source_relative,logtostderr=true:. --swagger_out=logtostderr=true:. $*.proto

# Handle any non-specific protobufs
%.pb.go: %.proto
	protoc ${PROTOBUF_INCLUDES} --gogoslick_out=paths=source_relative,plugins=grpc:. $*.proto

.PHONY: ${EXECUTABLE}
${EXECUTABLE}: tools ${PROTOS}
	# Compiling...
	go build -ldflags "-X ${PACKAGENAME}/conf.Executable=${EXECUTABLE} -X ${PACKAGENAME}/conf.GitVersion=${GITVERSION}" -o ${EXECUTABLE}

.PHONY: test
test: tools ${PROTOS}
	go test -cover ./...

deps:
	# Fetching dependancies...
	go get -d -v # Adding -u here will break CI

docker:
	docker build -t docker.io/snowzach/doods:local -f Dockerfile .

docker-images: docker-noavx docker-amd64 docker-arm32 docker-arm64
	docker manifest push --purge snowzach/doods:latest
	docker manifest create snowzach/doods:latest snowzach/doods:noavx snowzach/doods:arm32 snowzach/doods:arm64
	docker manifest push snowzach/doods:latest

.PHONY: docker-noavx
docker-noavx:
	docker build -t docker.io/snowzach/doods:noavx -f Dockerfile.noavx .
	docker push docker.io/snowzach/doods:noavx

.PHONY: docker-amd64
docker-amd64:
	docker build -t docker.io/snowzach/doods:amd64 -f Dockerfile.amd64 .
	docker push docker.io/snowzach/doods:amd64

.PHONY: docker-arm32
docker-arm32:
	docker build -t docker.io/snowzach/doods:arm32 -f Dockerfile.arm32 .
	docker push docker.io/snowzach/doods:arm32

.PHONY: docker-arm64
docker-arm64:
	docker build -t docker.io/snowzach/doods:arm64 -f Dockerfile.arm64 .
	docker push docker.io/snowzach/doods:arm64

.PHONY: docker-builder
docker-builder:
	docker build -t docker.io/snowzach/doods:builder -f Dockerfile.builder .

.PHONY: libedgetpu
libedgetpu:
	git clone https://github.com/google-coral/libedgetpu || true
	bash -c 'cd libedgetpu; DOCKER_CPUS="k8 armv7a aarch64" DOCKER_TARGETS=libedgetpu make docker-build'
