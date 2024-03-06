FROM golang:1.21.1-alpine3.18

RUN apk add --no-cache build-base curl protobuf ffmpeg

WORKDIR /go/src/app

COPY . .

# Install protoc compiler plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN go get -d -v ./...