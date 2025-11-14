FROM golang:1.24.9

ARG TUT_VERSION=0.1.0

RUN mkdir -p /app/configs
RUN mkdir -p /app/var/logs
RUN apt-get update

WORKDIR /app

RUN curl -sL https://github.com/Clivern/Tut/releases/download/v${TUT_VERSION}/tut_Linux_x86_64.tar.gz | tar xz
RUN rm LICENSE
RUN rm README.md

COPY ./config.dist.yml /app/configs/

EXPOSE 8080

VOLUME /app/configs
VOLUME /app/var

RUN ./tut version

CMD ["./tut", "server", "-c", "/app/configs/config.dist.yml"]
