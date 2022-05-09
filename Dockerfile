# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace

ARG go_proxy=https://goproxy.cn
ENV GO111MODULE=on \
    GOPROXY=${go_proxy} \
    CGO_ENABLED=0

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
# ENV GOPROXY https://goproxy.cn/
RUN go mod download


# Copy the go source
COPY main.go main.go

#COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN go build -ldflags '-w -s' -a -installsuffix cgo -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.azk8s.cn/distroless/static:latest
FROM alpine:3.15

RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk update && apk add subversion
RUN  echo 'hosts: files dns' > /etc/nsswitch.conf

WORKDIR /
COPY --from=builder /workspace/manager .
# USER nonroot:nonroot

ENTRYPOINT ["/manager"]
