# ベースイメージとして公式のGoイメージを使用
FROM golang:1.21.7-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# ホストのファイルをコンテナにコピー
COPY . .

# Goのビルドを実行
RUN go build -o main .

# 新しいステージを開始して最小限のイメージを作成
FROM alpine:latest

# # 作業ディレクトリを設定
WORKDIR /app

# ビルドしたバイナリを新しいコンテナにコピー
COPY --from=builder /app/main /app/main
COPY --from=builder /app/config.yml /app/config.yml

# コンテナが起動したときに実行するコマンドを設定
CMD ["/app/main"]
