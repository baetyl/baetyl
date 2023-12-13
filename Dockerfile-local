FROM alpine:3.12.3 as certs
RUN apk update && apk add ca-certificates

FROM --platform=$TARGETPLATFORM busybox
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
COPY baetyl /bin/
COPY /res/*.template /var/lib/baetyl/page/
ENTRYPOINT ["baetyl"]