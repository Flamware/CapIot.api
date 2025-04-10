# Stage 1: Build the Go application with Flyway
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk update && apk add --no-cache curl wget tar gzip

# Download and install Flyway
ARG FLYWAY_VERSION=9.23.1
ENV FLYWAY_HOME /flyway
ENV PATH ${PATH}:${FLYWAY_HOME}

RUN mkdir -p ${FLYWAY_HOME}/lib && \
    wget -qO- "https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}-linux-x64.tar.gz" | \
    tar -xzf - -C ${FLYWAY_HOME} --strip-components=1

# Create a directory for Flyway migrations
RUN mkdir -p /app/db/migration
COPY ../database/migration /app/db/migration

# ... (rest of your builder stage) ...

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy the built binary
COPY --from=builder /app/api ./api
# Copy Flyway CLI
COPY --from=builder /flyway/flyway ./flyway/flyway
# Copy migration scripts
COPY --from=builder /app/db/migration ./db/migration

RUN chmod +x ./flyway/flyway

# ... (rest of your runtime stage) ...