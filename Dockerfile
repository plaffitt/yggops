FROM golang:1.21-alpine3.18 as build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download -x

COPY cmd ./
ARG BINARY_NAME
RUN go build -o $BINARY_NAME
