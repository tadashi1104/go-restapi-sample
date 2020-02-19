# ベースとなるDockerイメージ指定
FROM golang:1.13
# コンテナ内に作業ディレクトリを作成
COPY ./ /go/src/example.com/t.szuuki/go-rest-sample/
# コンテナログイン時のディレクトリ指定
WORKDIR /go/src/example.com/t.szuuki/go-rest-sample
# ホストのファイルをコンテナの作業ディレクトリに移行
ADD . /go/src/example.com/t.szuuki/go-rest-sample

ENV GO111MODULE=on
RUN go mod download
EXPOSE 8080
EXPOSE 8000

CMD ["go", "run", "/go/src/example.com/t.szuuki/go-rest-sample/main.go"]
