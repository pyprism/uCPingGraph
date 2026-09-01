FROM golang:1.27-alpine

RUN apk add --no-cache inotify-tools gcc g++ musl-dev tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN go install github.com/air-verse/air@latest

ENV TZ=Asia/Dhaka

CMD air