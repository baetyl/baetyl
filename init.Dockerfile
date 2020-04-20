FROM --platform=$TARGETPLATFORM golang:1.13.5-stretch as devel
COPY / /go/src/
RUN cd /go/src/initialize/main && make all

FROM --platform=$TARGETPLATFORM busybox
COPY --from=devel /go/src/initialize/main/baetyl-init /bin/
COPY /initialize/res/*.template /var/lib/baetyl/page/
ENTRYPOINT ["baetyl-init"]