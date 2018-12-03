FROM python:2.7-alpine

RUN pip install --no-cache-dir protobuf

WORKDIR /home/openedge
COPY output/native/bin/openedge_hub ./bin/

ENTRYPOINT [ "bin/openedge_hub" ]
CMD []
