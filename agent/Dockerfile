FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o logsway-agent .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/logsway-agent .
COPY config.yaml.example config.yaml
ENTRYPOINT ["./logsway-agent", "-config", "/config/config.yaml"]
