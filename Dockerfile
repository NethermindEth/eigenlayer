FROM alpine:latest

ENV APP_NAME=eigen-wiz

WORKDIR /root/
COPY ./bin/${APP_NAME} ./${APP_NAME}
ENTRYPOINT ./${APP_NAME}