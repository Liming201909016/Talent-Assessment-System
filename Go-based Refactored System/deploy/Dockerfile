# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY configs /app/configs
EXPOSE 8092
ENTRYPOINT ["/app/server"]
