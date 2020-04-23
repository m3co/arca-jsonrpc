FROM golang:alpine

RUN apk add alpine-sdk

RUN mkdir -p /go/src/github.com/m3co/arca-jsonrpc/
WORKDIR /go/src/github.com/m3co/arca-jsonrpc/

COPY . .

CMD ["go", "test", "-v"]
