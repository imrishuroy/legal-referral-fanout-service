# Build stage
FROM golang:1.24-alpine3.20 AS builder

# Install necessary tools
RUN apk add --no-progress --no-cache git

WORKDIR /app

# Copy everything from the root directory into /app
COPY . .

# Build a static Linux/amd64 binary for reliability in minimal runtimes
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags "-s -w" -o main main.go

# Run state
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY service-account-key.json .

# Install CA certificates for outbound TLS (RDS sslmode=require, AWS SDK, etc.)
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Expose port 8080 for App Runner
EXPOSE 8080

# Specifies the executable command that runs when the container starts
CMD ["/app/main"]