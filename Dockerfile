FROM registry.bryk.io/general/shell:0.1.0

# Metadata
ARG VERSION
LABEL maintainer="Ben Cessa <ben@pixative.com>"
LABEL version=${VERSION}

# Expose required ports and volumes
EXPOSE 9090 9091

# Add application binary and use it as default entrypoint
COPY echo-server_linux_amd64 /bin/echo-server
ENTRYPOINT ["/bin/echo-server"]
