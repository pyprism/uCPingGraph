FROM golang:1.22.1-alpine3.18

RUN apk add --no-cache inotify-tools gcc g++ musl-dev

WORKDIR /app
COPY . /app

RUN go mod download

RUN go install github.com/cosmtrek/air@latest
CMD air