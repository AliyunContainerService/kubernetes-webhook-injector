## How to contribute 

## Build the project 

```cgo
# run make to build the project 
➜  kubernetes-webhook-injector git:(master) ✗ make
go fmt ./pkg/...
go vet ./pkg/...
go test ./pkg/...  -coverprofile cover.out
ok  	github.com/AliyunContainerService/kubernetes-webhook-injector/pkg	0.049s	coverage: 3.9% of statements
go build -o bin/kubernetes-webhook-injector main.go

# check the binary 
➜  kubernetes-webhook-injector git:(master) ✗ ls bin
kubernetes-webhook-injector
``` 

## Build the image 
```cgo
# run docker-build to build docker image 
➜  kubernetes-webhook-injector git:(master) ✗ make docker-build
go fmt ./pkg/...
go vet ./pkg/...
go test ./pkg/...  -coverprofile cover.out
ok  	github.com/AliyunContainerService/kubernetes-webhook-injector/pkg	0.026s	coverage: 3.9% of statements
docker build . -t registry.aliyuncs.com/acs/kubernetes-webhook-injector:v0.0.1-a931868-aliyun
Sending build context to Docker daemon  71.53MB
Step 1/9 : FROM golang:1.14.2 as builder
1.14.2: Pulling from library/golang
90fe46dd8199: Downloading  2.145MB
35a4f1977689: Downloading [========>                                          ]  883.1kB/5.487MB
bbc37f14aded: Downloading [===========>                                       ]  2.218MB/9.996MB
74e27dc593d4: Waiting
38b1453721cb: Waiting
780391780e20: Waiting
0f7fd9f8d114: Waiting

... 



```

## Debug the code 
```cgo

```