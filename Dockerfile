FROM golang:1.11 as builder

# ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
COPY pkg/ .
COPY cmd/ .

RUN go mod download
RUN go test ./... 
RUN go build -o bin/item/item ./cmd/item

FROM golang:1.11-alpine

COPY --from=builder /app/bin/item /usr/local/bin/