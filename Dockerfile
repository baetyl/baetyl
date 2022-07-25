FROM --platform=$TARGETPLATFORM golang:1.18.3-stretch as devel
ARG BUILD_ARGS
COPY / /go/src/
RUN cd /go/src/ && make build-local BUILD_ARGS=$BUILD_ARGS

FROM --platform=$TARGETPLATFORM busybox
ADD https://cacerts.digicert.com/DigiCertSHA2SecureServerCA.crt.pem /etc/ssl/certs/ca.crt
COPY --from=devel /go/src/baetyl /bin/
ENTRYPOINT ["baetyl"]