# Stage 1: Build the Go binary
FROM golang:1.24.1 AS builder

# Set working directory
WORKDIR /app

# Copy Go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary (disable CGO for smaller image)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/app .

# Expose the port your app runs on (e.g., 8080)
EXPOSE 8080

# Run the app
CMD ["./app"]