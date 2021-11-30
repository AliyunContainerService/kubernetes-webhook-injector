# Build the manager binary
FROM golang:1.17.3 as builder
# Copy in the go src
WORKDIR /go/src/github.com/AliyunContainerService/kubernetes-webhook-injector
COPY ./ /go/src/github.com/AliyunContainerService/kubernetes-webhook-injector
# Build
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/ && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build-binary

# Copy the controller-manager into a thin image
FROM debian:bullseye
RUN apt-get update \
    && apt-get install -y openssl curl tini \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /go/src/github.com/AliyunContainerService/kubernetes-webhook-injector/bin/kubernetes-webhook-injector .
COPY hack/gencerts.sh .

ENTRYPOINT  ["/root/kubernetes-webhook-injector"]
