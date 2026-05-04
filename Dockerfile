FROM alpine:3.21

RUN apk add --no-cache ca-certificates

# Create a non-root user
RUN addgroup -S litmus && adduser -S litmus -G litmus

# Copy the binary from the build context (GoReleaser puts it in the root)
COPY litmus /usr/local/bin/litmus

# Ensure the binary is executable
RUN chmod +x /usr/local/bin/litmus

USER litmus
WORKDIR /home/litmus

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/litmus"]
CMD ["check"]
