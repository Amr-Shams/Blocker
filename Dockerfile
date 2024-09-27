FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o blocker .
