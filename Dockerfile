# ----------------------------------------
# Stage 1: Build the Go binary
# ----------------------------------------
FROM golang:1.20-alpine AS builder

# Set working directory inside the builder
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first (caching)
COPY go.mod ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o s3fs-go

# ----------------------------------------
# Stage 2: Create minimal final image
# ----------------------------------------
FROM scratch

# Create and switch to /app as the working directory
WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/s3fs-go .

# Expose port 8080
EXPOSE 8080

# When the container starts, run the binary
ENTRYPOINT ["/app/s3fs-go"]

CMD [ "./storage" ]