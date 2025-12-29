FROM golang:1.25.2-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY chaincode.go ./

# Build the chaincode
RUN go build -o chaincode .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/chaincode .

# Expose chaincode port
EXPOSE 9999

# Run the chaincode
CMD ["./chaincode"]
