#!/bin/bash
export CGO_ENABLED=0
export GO111MODULE=on
export GOPROXY=https://goproxy.cn/
go mod tidy
echo "build amd64..."
GOARCH=amd64 go build -ldflags '-w -s' -a -installsuffix cgo -o bin/amd64/manager
echo "build arm64..."
GOARCH=arm64 go build -ldflags '-w -s' -a -installsuffix cgo -o bin/arm64/manager
 
# build and push images
docker buildx build --platform linux/amd64 -t index.alauda.cn/claas/captain-cert-init -f Dockerfile.init  . --push

#docker buildx build --platform linux/arm64 -t armharbor.alauda.cn/claas/captain-cert-init -f Dockerfile.init.arm . --push
docker buildx build --platform linux/arm64 -t  armharbor.alauda.cn/claas/captain -f Dockerfile.arch  . --push
docker buildx build --platform linux/amd64 -t  index.alauda.cn/claas/captain -f Dockerfile.arch  . --push

