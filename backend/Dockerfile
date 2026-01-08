# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install dependencies:
# - docker-cli: for registration via docker exec
# - tzdata: for Asia/Ho_Chi_Minh timezone support
RUN apk add --no-cache docker-cli tzdata

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 3000

# Set environment variables with defaults for single-org local fallback
ENV DB_HOST=localhost
ENV PEER_ENDPOINT=localhost:7051
ENV WALLET_PATH=/app/wallet
ENV CA1_URL=https://localhost:7054
ENV CA2_URL=https://localhost:8054
ENV IPFS_URL=localhost:5001

# Command to run
CMD ["./main"]
