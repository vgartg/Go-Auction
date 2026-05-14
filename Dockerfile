FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o goauction ./cmd/goauction

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/goauction .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./goauction"]