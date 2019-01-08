FROM golang:alpine as builder
RUN apk add --update --no-cache make
COPY openedge /go/src/github.com/baidu/openedge
RUN cd /go/src/github.com/baidu/openedge/openedge-remote-mqtt && make rebuild

FROM alpine
COPY --from=builder /go/src/github.com/baidu/openedge/openedge-remote-mqtt /bin/
ENTRYPOINT ["openedge-remote-mqtt"]
CMD []
