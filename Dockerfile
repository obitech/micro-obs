FROM golang:1.11 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum Makefile ./
COPY item/ item/
COPY order/ order/
COPY cmd/ cmd/
COPY util/ util/

RUN go mod download
RUN make build
FROM alpine:latest

COPY --from=builder /app/bin/* /usr/local/bin/