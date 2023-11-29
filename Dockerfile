# Build
FROM golang:1.21-alpine3.18 as build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download -x

COPY cmd ./
RUN go build -o generic-gitops

# Final image
FROM alpine:3.18

RUN apk add --no-cache git openssh python3 docker docker-cli-compose

WORKDIR /

COPY --from=build /build/generic-gitops /usr/local/bin

COPY plugins /var/lib/generic-gitops/plugins

CMD ["generic-gitops"]
