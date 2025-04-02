# Stage 1: Build the Go application
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/api ./api

# Expose the port your API listens on
EXPOSE 8080

# Command to run the executable
CMD ["./api"]