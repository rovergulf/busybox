#syntax=docker/dockerfile:1.2
FROM golang:alpine AS build

RUN apk --update add ca-certificates git

ARG APP_VERSION

WORKDIR /build

COPY . .

RUN go mod tidy
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags "-X github.com/rovergulf/busybox/handler.AppVersion=$APP_VERSION" -o busybox .

FROM alpine

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

COPY --from=build /build/busybox /app

ENTRYPOINT ["/app/busybox"]
