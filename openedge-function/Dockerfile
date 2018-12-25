FROM golang:alpine as builder
RUN apk add --update --no-cache make
COPY openedge /go/src/github.com/baidu/openedge
WORKDIR /go/src/github.com/baidu/openedge/openedge-function
RUN make rebuild

FROM alpine
COPY --from=builder /go/src/github.com/baidu/openedge/openedge-function /bin/
ENTRYPOINT ["openedge-function"]
CMD []
