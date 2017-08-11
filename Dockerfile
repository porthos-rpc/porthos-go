FROM golang:1.8

RUN apt-get update && apt-get install -y wget
RUN wget https://github.com/jwilder/dockerize/releases/download/v0.2.0/dockerize-linux-amd64-v0.2.0.tar.gz
RUN tar -C /usr/local/bin -xzvf dockerize-linux-amd64-v0.2.0.tar.gz

RUN mkdir -p /go/src/github.com/porthos-rpc/porthos-go
WORKDIR /go/src/github.com/porthos-rpc/porthos-go

RUN go get github.com/streadway/amqp

ADD . /go/src/github.com/porthos-rpc/porthos-go