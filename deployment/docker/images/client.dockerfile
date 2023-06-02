FROM golang:1.16.3-alpine3.13 as builder

RUN apk add --no-cache \
	bash \
	make \
    gcc \
    musl-dev

WORKDIR /client
ADD . /client/
COPY ./env.secret /client/env

RUN make tidy
RUN make build-client

FROM alpine:latest

ENV TRANSACTION_SERVER_PORT=9000 PROFILING_SERVER_PORT=6060

WORKDIR /client
COPY --from=builder /client/build/client_linux/clientchain .

EXPOSE $TRANSACTION_SERVER_PORT $PROFILING_SERVER_PORT
CMD ["./clientchain"]
