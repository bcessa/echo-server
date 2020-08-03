# Echo Server

[![Build Status](https://github.com/bcessa/echo-server/workflows/ci/badge.svg?branch=master)](https://github.com/bcessa/echo-server/actions)
[![Version](https://img.shields.io/github/tag/bcessa/echo-server.svg)](https://github.com/bcessa/echo-server/releases)
[![Software License](https://img.shields.io/badge/license-BSD3-red.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/bcessa/echo-server?style=flat)](https://goreportcard.com/report/github.com/bcessa/echo-server)

This project provides a sample service intended only to run tests of deployment and ingress setup.

When exposing the service as HTTP only the SSL termination can be provided by the ingress. When
exposing the gRPC port through a load balancer only secure connections are supported and the SSL
must be terminated by the pod. This requires the load balancer, ingress controller and ingress
resource to be properly configured to support __SSL passthroughs__.

When using this chart in this way the value `conf.certificate` is required. It must be set to a
secret containing the TLS certificate. The certificate will be used by the Pod to terminate TLS,
__NOT__ the ingress resource.
[TLS Secret Reference](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls) 

More information:

- Load Balancer [DigitalOcean](https://github.com/digitalocean/digitalocean-cloud-controller-manager/blob/master/docs/controllers/services/annotations.md#servicebetakubernetesiodo-loadbalancer-tls-passthrough)
- Ingress Controller [Nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/user-guide/tls/#ssl-passthrough)
- Ingress Resource [Nginx Ingress Resource](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#ssl-passthrough)
