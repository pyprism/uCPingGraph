FROM golang:1.20.6-alpine3.17

RUN apk add --no-cache inotify-tools gcc g++

WORKDIR /app
COPY . /app

RUN go mod download

RUN go install github.com/cosmtrek/air@latest
CMD air