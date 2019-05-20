FROM scratch

ARG VERSION
LABEL maintainer="Ben Cessa <ben@pixative.com>"
LABEL version=${VERSION}

COPY echo-server-linux /
COPY ca-roots.crt /etc/ssl/certs/

EXPOSE 9090

ENTRYPOINT ["/echo-server-linux"]
