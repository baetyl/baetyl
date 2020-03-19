FROM --platform=$TARGETPLATFORM golang:1.13.5-stretch as devel
COPY / /go/src/
RUN cd /go/src/ && make build-static

FROM --platform=$TARGETPLATFORM busybox
COPY --from=devel /go/src/baetyl-core /bin/
ENTRYPOINT ["baetyl-core"]