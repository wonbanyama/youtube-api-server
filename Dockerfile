FROM golang:1.24 AS builder

RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    git \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .env ./

COPY server/ ./server/

RUN go build -o /app/app ./server

FROM ubuntu:24.04

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 3000

CMD ["./app"]
