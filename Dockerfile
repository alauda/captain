FROM golang:1.12.4

COPY . $GOPATH/src/alauda.io/captain
WORKDIR $GOPATH/src/alauda.io/captain
RUN make build

FROM index.alauda.cn/alaudaorg/alaudabase-alpine-run:alpine3.9.3

WORKDIR /captain

COPY --from=0 /go/src/alauda.io/captain/captain /captain/
COPY hack/run.sh /captain/run.sh
RUN chmod a+x /captain/captain /captain/run.sh

ENTRYPOINT ["/captain/run.sh"]
CMD ["/captain/captain"]