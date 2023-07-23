FROM golang:1.20.6-alpine3.17

RUN apk add --no-cache inotify-tools

WORKDIR /app
COPY . /app

RUN go mod download

RUN go install github.com/cosmtrek/air@latest
CMD air