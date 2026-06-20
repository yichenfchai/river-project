# Grand Canal Guardian — 单体应用 Dockerfile
# 多阶段构建：build → run

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/server /server
COPY --from=builder /app/web/dist /web/dist
EXPOSE 8080
ENTRYPOINT ["/server"]
