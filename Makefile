# build params
PREFIX?=registry.aliyuncs.com/acs
VERSION?=v0.0.1
GIT_COMMIT:=$(shell git rev-parse --short HEAD)

# Image URL to use all building/pushing image targets
IMG ?= $(PREFIX)/kubernetes-webhook-injector:$(VERSION)-$(GIT_COMMIT)-aliyun
all: test kubernetes-webhook-injector

# Run tests
test: fmt vet
	go test ./pkg/...  -coverprofile cover.out

# Build kubernetes-webhook-injector binary
kubernetes-webhook-injector: fmt vet
	go build -o bin/kubernetes-webhook-injector github.com/AliyunContainerService/kubernetes-webhook-injector/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run ./main.go

# Run go fmt against code
fmt:
	go fmt ./pkg/...

# Run go vet against code
vet:
	go vet ./pkg/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG} -f deploy/Dockerfile

# Push the docker image
docker-push:
	docker push ${IMG}