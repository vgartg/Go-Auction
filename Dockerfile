FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY . .
RUN templ generate
RUN go build -o goauction ./cmd/goauction

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/goauction .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./goauction"]
