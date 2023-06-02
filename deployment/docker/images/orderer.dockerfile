FROM golang:1.16.3-alpine3.13 as builder

RUN apk add --no-cache \
	bash \
	make

WORKDIR /orderer
ADD . /orderer
COPY ./env.secret /orderer/env

RUN make tidy
RUN make build-orderer

FROM alpine:latest

ENV ORDERER_SERVER_PORT=9000 PROFILING_SERVER_PORT=6060

WORKDIR /orderer
COPY --from=builder /orderer/build/orderer_linux/ordererchain .

EXPOSE $ORDERER_SERVER_PORT $PROFILING_SERVER_PORT
CMD ["./ordererchain"]
