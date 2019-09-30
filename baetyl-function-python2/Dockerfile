FROM --platform=$TARGETPLATFORM python:2-slim-buster
COPY *.py *.txt /bin/
RUN cd /bin/ && chmod +x function-python2.py && pip install -r requirements.txt
ENTRYPOINT ["function-python2.py"]