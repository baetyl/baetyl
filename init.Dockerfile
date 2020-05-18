FROM --platform=$TARGETPLATFORM golang:1.13.5-stretch as devel
COPY / /go/src/
RUN cd /go/src/ && make init

FROM --platform=$TARGETPLATFORM busybox
COPY --from=devel /go/src/baetyl-init /bin/
COPY /initz/res/ /var/lib/baetyl/page/
ENTRYPOINT ["baetyl-init"]