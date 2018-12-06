FROM golang:1.11 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
COPY item/ item/
COPY cmd/ cmd/
COPY util/ util/

RUN go mod download
RUN go test ./... 
RUN go build -o bin/item ./cmd/item

FROM alpine:latest

COPY --from=builder /app/bin/item /usr/local/bin/