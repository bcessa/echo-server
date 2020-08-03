.PHONY: *
default: help
VERSION_TAG=0.1.2
BINARY_NAME=echo-server
DOCKER_IMAGE_NAME=docker.pkg.github.com/bcessa/echo-server/echo-server

# Linker tags
# https://golang.org/cmd/link/
LD_FLAGS += -s -w
LD_FLAGS += -X github.com/bcessa/echo-server/cmd.coreVersion=$(VERSION_TAG)
LD_FLAGS += -X github.com/bcessa/echo-server/cmd.buildTimestamp=$(shell date +'%s')
LD_FLAGS += -X github.com/bcessa/echo-server/cmd.buildCode=$(shell git log --pretty=format:'%H' -n1)

## help: Prints this help message
help:
	@echo "Commands available"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /' | sort

## clean: Download and compile all dependencies and intermediary products
clean:
	@-rm -rf vendor
	go mod tidy
	go mod verify
	go mod download
	go mod vendor

## updates: List available updates for direct dependencies
# https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies
updates:
	@go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all 2> /dev/null

## scan: Look for known vulnerabilities in the project dependencies
# https://github.com/sonatype-nexus-community/nancy
scan:
	@nancy -quiet go.sum

## lint: Static analysis
lint:
	helm lint helm/*
	golangci-lint run -v ./...

## test: Run unit tests excluding the vendor dependencies
test:
	go test -race -cover -v -failfast ./...

## ca-roots: Generate the list of valid CA certificates
ca-roots:
	@docker run -dit --rm --name ca-roots debian:stable-slim
	@docker exec --privileged ca-roots sh -c "apt update"
	@docker exec --privileged ca-roots sh -c "apt install -y ca-certificates"
	@docker exec --privileged ca-roots sh -c "cat /etc/ssl/certs/* > /ca-roots.crt"
	@docker cp ca-roots:/ca-roots.crt ca-roots.crt
	@docker stop ca-roots

## build: Build for the default architecture in use
build:
	go build -v -ldflags '$(LD_FLAGS)' -o $(BINARY_NAME)

## build-for: Build the available binaries for the specified 'os' and 'arch'
# make build-for os=linux arch=amd64
build-for:
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) \
	go build -v -ldflags '$(LD_FLAGS)' \
	-o $(dest)$(BINARY_NAME)_$(VERSION_TAG)_$(os)_$(arch)$(suffix)

## docker: Build docker image
docker:
	make build-for os=linux arch=amd64
	@-docker rmi $(DOCKER_IMAGE_NAME):$(VERSION_TAG)
	@docker build --build-arg VERSION="$(VERSION_TAG)" --rm -t $(DOCKER_IMAGE_NAME):$(VERSION_TAG) .

## release: Prepare artifacts for a new tagged release
release:
	@-rm -rf release-$(VERSION_TAG)
	mkdir release-$(VERSION_TAG)
	make build-for os=linux arch=amd64 dest=release-$(VERSION_TAG)/
	make build-for os=darwin arch=amd64 dest=release-$(VERSION_TAG)/
	make build-for os=windows arch=amd64 suffix=".exe" dest=release-$(VERSION_TAG)/
	make build-for os=windows arch=386 suffix=".exe" dest=release-$(VERSION_TAG)/

## ci-update: Update the signature on the CI configuration file
ci-update:
	drone sign bcessa/echo-server --save
