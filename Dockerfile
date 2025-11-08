# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies including swag for Swagger generation
RUN apk add --no-cache git && \
    go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger documentation
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gocrawl .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/gocrawl .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./gocrawl"]
