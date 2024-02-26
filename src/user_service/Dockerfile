FROM golang:1.21.1-alpine3.18

RUN apk add --no-cache build-base curl protobuf

RUN curl https://downloads.sqlc.dev/sqlc_1.22.0_linux_amd64.tar.gz --output sqlc.tar.gz
RUN tar -xvzf sqlc.tar.gz
RUN mv sqlc bin/

RUN curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh

WORKDIR /go/src/app

COPY . .

# Install protoc compiler plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN go get -d -v ./...