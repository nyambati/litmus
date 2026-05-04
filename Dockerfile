# Stage 1: Build the UI
FROM node:20-alpine AS ui-builder
WORKDIR /app/ui
# Copy package files first for better caching
COPY ui/package*.json ./
RUN npm ci --prefer-offline
# Copy UI source and build
COPY ui/ ./
RUN npm run build

# Stage 2: Build the Go binary
FROM golang:1.26-alpine AS go-builder
ARG VERSION=dev
WORKDIR /app
# Install git for potential private dependencies and ca-certificates
RUN apk add --no-cache git ca-certificates
# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download
# Copy everything else
COPY . .
# Copy the built UI dist from the previous stage (it's embedded by Go)
COPY --from=ui-builder /app/ui/dist ./ui/dist
# Build the binary with the injected version
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o litmus .

# Stage 3: Final lightweight image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
# Create a non-root user
RUN addgroup -S litmus && adduser -S litmus -G litmus
USER litmus
WORKDIR /home/litmus
# Copy binary from builder
COPY --from=go-builder /app/litmus /usr/local/bin/litmus
# Default entrypoint
ENTRYPOINT ["litmus"]
CMD ["check"]
