# Stage 1: Build the Go application with Flyway
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk update && apk add --no-cache curl wget tar gzip

# Download and install Flyway
ARG FLYWAY_VERSION=9.23.1 # You can adjust the version
ENV FLYWAY_HOME /flyway
ENV PATH ${PATH}:${FLYWAY_HOME}

RUN mkdir -p ${FLYWAY_HOME}/lib && \
    wget -qO- "https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}-linux-x64.tar.gz" | \
    tar -xzf - -C ${FLYWAY_HOME} --strip-components=1

# Create a directory for Flyway migrations INSIDE the /app directory
RUN mkdir -p /app/db/migration

# Copy migration scripts from the PARENT directory
COPY ../database/migration /app/db/migration

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN apk update && apk add curl # You had this twice, keeping one

# Copy the application source code
COPY .. .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/server.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/api ./api
COPY --from=builder /app/db/migration ./db/migration # Copy migration scripts
COPY --from=builder /flyway/flyway ./flyway/flyway # Copy Flyway CLI

# Make Flyway executable
RUN chmod +x ./flyway/flyway

# Expose the port your API listens on
EXPOSE 8080

# Command to run the executable
CMD ["./api"]