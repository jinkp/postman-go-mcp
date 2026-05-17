# Stage 1: build
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mcp-postman ./cmd/server

# Stage 2: runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
RUN mkdir -p /app/data

COPY --from=builder /app/mcp-postman .

EXPOSE 8080
ENTRYPOINT ["./mcp-postman"]
