FROM golang:1.12.9

COPY . $GOPATH/src/github.com/alauda/captain
WORKDIR $GOPATH/src/github.com/alauda/captain
RUN make build

FROM index.alauda.cn/alaudaorg/alaudabase-alpine-run:alpine3.10

WORKDIR /captain

COPY --from=0 /go/src/github.com/alauda/captain/captain /captain/
COPY artifacts/helm/repositories.yaml /root/.config/helm/
RUN chmod a+x /captain/captain


# ENTRYPOINT ["/captain/run.sh"]
CMD ["/captain/captain"]
