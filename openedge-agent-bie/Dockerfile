FROM golang:alpine as builder
RUN apk add --update --no-cache make
COPY openedge /go/src/github.com/baidu/openedge
RUN cd /go/src/github.com/baidu/openedge/openedge-agent-bie && make rebuild

FROM alpine
COPY --from=builder /go/src/github.com/baidu/openedge/openedge-agent-bie /bin/
ENTRYPOINT ["openedge-agent-bie"]
CMD []
