FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 go build -o /go-chess

# Stage 2: Run
# Use a minimal base image like 'scratch' or 'alpine' for the final runtime image
FROM scratch 

# Copy the compiled binary from the builder stage
COPY --from=builder /go-chess /go-chess
# COPY --from=builder /app/public /app/public # Copy static assets if any

# Expose the port the Gin app listens on (default is 8080)
EXPOSE 8080

# Set environment variable for Gin release mode
ENV GIN_MODE release

# Command to run the executable
CMD ["/go-chess"]
