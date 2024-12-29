FROM golang:1.22.9-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o tbls-ask-bot .


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/tbls-ask-bot /app/tbls-ask-bot

ENTRYPOINT ["/app/tbls-ask-bot"]

CMD ["server"]
