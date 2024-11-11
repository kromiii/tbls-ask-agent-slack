FROM golang:1.22.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o tbls-ask-bot .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/tbls-ask-bot /app/tbls-ask-bot

RUN apk add --no-cache curl

ENTRYPOINT ["/app/tbls-ask-bot"]

CMD ["server"]

