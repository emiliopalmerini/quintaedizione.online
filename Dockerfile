# Build stage
FROM golang:1.25-alpine AS builder

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install Templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
COPY . .

# Generate Templ templates
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/app

# Ensure required directories exist
RUN mkdir -p configs

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/app .

# Copy web assets
COPY --from=builder /app/web ./web

# Copy configs directory
COPY --from=builder /app/configs ./configs

# Expose port
EXPOSE 8000

# Run the binary
CMD ["./app"]
