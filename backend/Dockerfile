FROM golang:latest

WORKDIR /app

# ローカルのソースコードをコンテナにコピー
COPY . .

# 必要なパッケージをインストール
RUN go mod download

# アプリケーションをビルド
RUN go build -o main .

# コンテナのポートを公開
EXPOSE 8080

# アプリケーションを実行
CMD ["go", "run", "main.go"]