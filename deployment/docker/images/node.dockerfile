FROM golang:1.16.3-alpine3.13 as builder

RUN apk add --no-cache \
	bash \
	make

WORKDIR /node
ADD . /node/
COPY ./env.secret /node/env

RUN make tidy
RUN make build-node

FROM alpine:latest

ENV TRANSACTION_SERVER_PORT=9000 PROFILING_SERVER_PORT=6060

WORKDIR /node
COPY --from=builder /node/build/node_linux/nodechain .

EXPOSE $TRANSACTION_SERVER_PORT $PROFILING_SERVER_PORT
CMD ["./nodechain"]
