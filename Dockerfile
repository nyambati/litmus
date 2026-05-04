FROM alpine:3.21
RUN apk add --no-cache ca-certificates
# Create a non-root user
RUN addgroup -S litmus && adduser -S litmus -G litmus
USER litmus
WORKDIR /home/litmus

ARG TARGETPLATFORM
# GoReleaser places the binary in a directory named after the platform
COPY $TARGETPLATFORM/litmus /usr/local/bin/litmus

# Default entrypoint
ENTRYPOINT ["litmus"]
CMD ["check"]
