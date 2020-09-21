FROM alpine:3.11

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk upgrade --update-cache --available && \
    apk add openssl && \
    rm -rf /var/cache/apk/*

COPY generate_certificate.sh /

COPY --from=lachlanevenson/k8s-kubectl:v1.19.2 /usr/local/bin/kubectl /usr/local/bin/kubectl
