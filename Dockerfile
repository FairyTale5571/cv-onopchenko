# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Generate templ files
RUN templ generate

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Add CA certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D appuser

# Copy the binary and required files from builder
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/static ./static

# Set proper permissions
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Set environment variables
ENV TZ=UTC \
    PORT=8080

# Run the binary
CMD ["./main"] 