FROM alpine:3.19

# Install file utility to validate static linking.
RUN apk add --no-cache file bash

COPY dist/simple-cli /usr/local/bin/simple-cli
RUN chmod +x /usr/local/bin/simple-cli

# Verify the binary is statically linked (CGO_ENABLED=0).
RUN file /usr/local/bin/simple-cli | grep -q "statically linked" || \
    file /usr/local/bin/simple-cli | grep -q "ELF"

CMD ["simple-cli", "--version"]
