FROM golang:1.23.3-alpine3.20 AS builder

RUN apk add --no-progress --no-cache gcc musl-dev

WORKDIR /app

# Copies everything from your root directory into /app
COPY . .

# Builds your app with optional configuration
RUN go build -tags 'lambda.norpc musl' -ldflags '-extldflags "-static"' -o main main.go

# Run state
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .

# Specifies the executable command that runs when the container starts
CMD ["/app/main"]