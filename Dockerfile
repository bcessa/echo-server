FROM busybox:1.31.1

ARG VERSION
LABEL maintainer="Ben Cessa <ben@pixative.com>"
LABEL version=${VERSION}

COPY echo-service-linux /bin
COPY ca-roots.crt /etc/ssl/certs/

EXPOSE 9090 9091

ENTRYPOINT ["/bin/echo-service-linux"]
