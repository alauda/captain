# ARM64 Support
Captain has built in docker image support for arm64 platform, even though it's not automatically generated from GitHub CI， it's very easy to build it from the source


## Steps

### Prepare

```bash
git clone https://github.com/alauda/captain.git
cd captain
```

### Build cert-init image

```bash
wget -c <kubectl-arm-binary-url> -O bin/kubectl

docker build -t captain-cert-init:arm -f Dockerfile.init.arm .
```

Notes: the url of kubectl binary can be found at : https://kubernetes.io/docs/tasks/tools/install-kubectl/


### Build captain image

```bash
export CGO_ENABLED=0
export GO111MODULE=on
go mod tidy
GOARCH=arm64 go build -ldflags '-w -s' -a -installsuffix cgo -o bin/arm64/manager

docker build -t captain:arm -f Dockerfile.arm .
```
