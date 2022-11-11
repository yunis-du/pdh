FROM golang:alpine as builder

RUN mkdir /var/log/ultraflow/

WORKDIR /pdh/relay
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o pdh-relay .

FROM alpine:latest

LABEL maintainer=duyunzhi<duyzmail@163.com>\
description="pdh relay"

RUN echo -e http://mirrors.ustc.edu.cn/alpine/latest-stable/main/ > /etc/apk/repositories && apk update && apk add tzdata

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

WORKDIR /pdh/relay

COPY --from=0 /pdh/relay/pdh-relay ./

ENTRYPOINT ./pdh-relay relay --port 6880