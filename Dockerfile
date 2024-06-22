FROM golang:1.22

WORKDIR /app

# Copy go.mod and go.sum to the WORKDIR
COPY go.mod go.sum ./

# Copies everything from your root directory into /app
COPY . .

# Installs Go dependencies
RUN go mod download

# Builds your app with optional configuration
RUN go build -o /main

# Tells Docker which network port your container listens on
EXPOSE 8080

# Specifies the executable command that runs when the container starts
CMD ["/main"]