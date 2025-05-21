FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# Build arguments for multi-platform support
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Set default values in case they're not passed
ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build for the target architecture
RUN echo "Building for OS=${TARGETOS} ARCH=${TARGETARCH}"
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o stock-ticker cmd/main.go

# Use a small alpine image for the final container
FROM --platform=$TARGETPLATFORM alpine:latest

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set up the working directory
WORKDIR /root/

# Copy only the binary from the builder stage
COPY --from=builder /app/stock-ticker .

# Expose the port the server listens on
EXPOSE 8080

# Run the binary
CMD ["./stock-ticker"]