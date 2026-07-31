FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o tbls-ask-bot .


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/tbls-ask-bot /app/tbls-ask-bot
COPY --from=builder /app/LICENSE /app/LICENSE
COPY --from=builder /app/CREDITS /app/CREDITS

ENTRYPOINT ["/app/tbls-ask-bot"]

CMD ["server"]
