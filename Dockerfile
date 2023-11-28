FROM ubuntu:20.04
LABEL authors="zq"

COPY webook /app/webook
WORKDIR /app

ENTRYPOINT ["/app/webook"]