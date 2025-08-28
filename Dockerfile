# Build stage
FROM golang:1.19 as build

WORKDIR /app

# Copy go.mod and go.sum first for efficient dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o garmin-connect ./cmd/main.go

# Runtime stage
FROM alpine:3.14

# Install CA certificates for SSL
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/garmin-connect .

# Set entrypoint
ENTRYPOINT ["./garmin-connect"]
