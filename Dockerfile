FROM docker.pkg.github.com/bryk-io/base-images/shell:0.1.0

# Expose required ports and volumes
EXPOSE 9090 9091

# Add application binary and use it as default entrypoint
COPY echo-server_linux_amd64 /bin/echo-server
ENTRYPOINT ["/bin/echo-server"]
