FROM golang:1.21-alpine3.18 as build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download -x

COPY cmd ./cmd
COPY internal ./internal

ARG BINARY_NAME
RUN go build -o $BINARY_NAME -v ./cmd/main.go
