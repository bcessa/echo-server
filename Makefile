.PHONY: *
default: help

# Project setup
BINARY_NAME=echo-server
DOCKER_IMAGE_NAME=docker.pkg.github.com/bcessa/echo-server/echo-server
MAINTAINERS='Ben Cessa <ben@pixative.com>'

# State values
GIT_COMMIT_DATE=$(shell TZ=UTC git log -n1 --pretty=format:'%cd' --date='format-local:%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT_HASH=$(shell git log -n1 --pretty=format:'%H')
GIT_TAG=$(patsubst v%,%,$(shell git describe --abbrev=0 --match='v*' --always | cut -c 1-8))

# Linker tags
# https://golang.org/cmd/link/
LD_FLAGS += -s -w
LD_FLAGS += -X github.com/bcessa/echo-server/cmd.coreVersion=$(GIT_TAG)
LD_FLAGS += -X github.com/bcessa/echo-server/cmd.buildTimestamp=$(GIT_COMMIT_DATE)
LD_FLAGS += -X github.com/bcessa/echo-server/cmd.buildCode=$(GIT_COMMIT_HASH)

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
	go test -race -v -failfast -coverprofile=coverage.report ./...
	go tool cover -html coverage.report -o coverage.html

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
	-o $(BINARY_NAME)_$(os)_$(arch)$(suffix)

## release: Prepare artifacts for a new tagged release
release:
	goreleaser release --skip-validate --skip-publish --rm-dist

## docker: Build docker image
# https://github.com/opencontainers/image-spec/blob/master/annotations.md
docker:
	make build-for os=linux arch=amd64
	@-docker rmi $(DOCKER_IMAGE_NAME):$(GIT_TAG)
	@docker build \
	"--label=org.opencontainers.image.title=echo-server" \
	"--label=org.opencontainers.image.authors=$(MAINTAINERS)" \
	"--label=org.opencontainers.image.created=$(GIT_COMMIT_DATE)" \
	"--label=org.opencontainers.image.revision=$(GIT_COMMIT_HASH)" \
	"--label=org.opencontainers.image.version=$(GIT_TAG)" \
	--rm -t $(DOCKER_IMAGE_NAME):$(GIT_TAG) .
	@rm $(BINARY_NAME)_linux_amd64
