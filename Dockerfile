FROM golang:1.22.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/main

CMD ["/app/main"]
