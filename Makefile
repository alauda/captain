VERSION := $(shell git describe --always --long --dirty)
mod:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

build:
	GO111MODULE=on CGO_ENABLED=0 go build -mod vendor -ldflags "-w -s -X main.version=${VERSION}" -v -o captain


fmt:
	find ./pkg -name \*.go  | xargs goimports -w
	goimports -w main.go

lint:
	golangci-lint run -c hack/ci/golangci.yml
	revive -exclude pkg/apis/... -exclude pkg/client/... -config hack/ci/revive.toml -formatter friendly ./pkg/...

test:
	go test  -v -cover -coverprofile=artifacts/coverage.out ./pkg/...

image:
	DOCKER_BUILDKIT=1 docker build -t captain .

push:
	docker tag captain index.alauda.cn/alaudaorg/captain:latest
	docker push index.alauda.cn/alaudaorg/captain:latest

code-gen:
	${GOPATH}/src/k8s.io/code-generator/generate-groups.sh all "alauda.io/captain/pkg/client" "alauda.io/captain/pkg/apis" app:v1alpha1

check: fmt build lint test
