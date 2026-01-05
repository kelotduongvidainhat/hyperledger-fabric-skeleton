# Stage 1: Build the binary
FROM golang:1.20-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum first to leverage layer caching
COPY go.mod go.sum ./
# Vendor dependencies are already there if 'go mod vendor' was run, but 'go mod download' is safer if not
RUN go mod download

COPY . .

# Build the binary
RUN go build -o chaincode -v .

# Stage 2: Run the binary
FROM alpine:3.17

WORKDIR /app

COPY --from=build /app/chaincode /app/chaincode

# Expose the port for the chaincode server
EXPOSE 9999

# Start command (will be overridden or used as default)
CMD ["./chaincode"]
