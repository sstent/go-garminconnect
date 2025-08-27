# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/go-garminconnect ./cmd

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/go-garminconnect .
EXPOSE 8080
CMD ["./go-garminconnect"]
