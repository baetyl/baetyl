FROM --platform=$TARGETPLATFORM golang:1.20 as devel
ARG BUILD_ARGS
COPY / /go/src/
RUN cd /go/src/ && make build-local BUILD_ARGS=$BUILD_ARGS

FROM alpine:3.12.3 as certs
RUN apk update && apk add ca-certificates

FROM --platform=$TARGETPLATFORM busybox
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
COPY --from=devel /go/src/baetyl /bin/
COPY /res/*.template /var/lib/baetyl/page/
ENTRYPOINT ["baetyl"]