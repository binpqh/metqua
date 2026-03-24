FROM debian:12-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY dist/simple-cli /usr/local/bin/simple-cli
RUN chmod +x /usr/local/bin/simple-cli

CMD ["simple-cli", "--version"]
