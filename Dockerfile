FROM alpine:latest

ENV APP_NAME=eigenlayer

WORKDIR /root/
COPY ./bin/${APP_NAME} ./${APP_NAME}
ENTRYPOINT ./${APP_NAME}