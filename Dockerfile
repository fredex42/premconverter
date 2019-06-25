FROM alpine:latest

COPY bin/premconverter.linux64 /usr/local/bin/premconverter
RUN chmod a+x /usr/local/bin/premconverter
USER daemon