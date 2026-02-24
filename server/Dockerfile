FROM golang:1.26-alpine

RUN apk add --no-cache inotify-tools gcc g++ musl-dev tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN go install github.com/cosmtrek/air@latest

ENV TZ=Asia/Dhaka

CMD air