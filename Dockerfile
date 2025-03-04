# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ovh-checker .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ovh-checker .

# Set the entrypoint
CMD ["/app/ovh-checker"]
