FROM golang:1.22.9-alpine AS builder

WORKDIR /app

COPY . .

RUN apk add --no-cache gcc musl-dev

RUN CGO_ENABLED=1 go build -o tbls-ask-bot .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/tbls-ask-bot /app/tbls-ask-bot

RUN apk add --no-cache curl sqlite-libs

ENTRYPOINT ["/app/tbls-ask-bot"]

CMD ["server"]
