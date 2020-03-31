FROM busybox:1.31.1

# Metadata
ARG VERSION
LABEL maintainer="Ben Cessa <ben@pixative.com>"
LABEL version=${VERSION}

# Install trusted certificate roots
COPY ca-roots.crt /etc/ssl/certs/

# Run as an unprivileged user
ENV USER=guest
ENV UID=10001
RUN adduser -h /home/${USER} -g "container-user" -s /bin/sh -D -u 10001 ${USER} ${USER}
USER ${USER}:${USER}

# Expose required ports and volumes
EXPOSE 9090 9091

# Add application binary and use it as default entrypoint
COPY echo-service-linux /bin
ENTRYPOINT ["/bin/echo-service-linux"]
