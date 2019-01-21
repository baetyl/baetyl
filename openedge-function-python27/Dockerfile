FROM python:2.7-alpine
WORKDIR /
COPY openedge-function-python27 runtime /bin/
RUN chmod +x /bin/runtime
ENTRYPOINT ["/bin/openedge-function-python27"]
