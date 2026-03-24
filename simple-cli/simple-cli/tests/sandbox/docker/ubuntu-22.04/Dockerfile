FROM ubuntu:22.04

# Install curl for downloading the binary if not volume-mounted.
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the pre-built binary from the build context (dist/).
COPY dist/simple-cli /usr/local/bin/simple-cli
RUN chmod +x /usr/local/bin/simple-cli

CMD ["simple-cli", "--version"]
