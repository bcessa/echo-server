.PHONY: proto
default: help
DOCKER_IMAGE_NAME=bcessa/echo-service
BINARY_NAME=echo-service
VERSION_TAG=0.1.0

# Linker tags
# https://golang.org/cmd/link/
LD_FLAGS="\
-X 'github.com/bcessa/echo-server/cmd.coreVersion=$(VERSION_TAG)' \
-X 'github.com/bcessa/echo-server/cmd.buildTimestamp=`date +'%s'`' \
-X 'github.com/bcessa/echo-server/cmd.buildCode=`git log --pretty=format:'%H' -n1`' \
"

help: ## Display available make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[33m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## Download and compile all dependencies and intermediary products
	@-rm -rf vendor
	go mod tidy
	go mod verify
	go mod download
	go mod vendor

updates: ## List available updates for direct dependencies
	# https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies
	go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all 2> /dev/null

build: ## Build for the default architecture in use
	go build -v -ldflags $(LD_FLAGS) -o $(BINARY_NAME)

linux: ## Build for linux systems
	GOOS=linux GOARCH=amd64 go build -v -ldflags $(LD_FLAGS) -o $(BINARY_NAME)-linux

docker: ## Build docker image
	make linux
	@-docker rmi $(DOCKER_IMAGE_NAME):$(VERSION_TAG)
	@docker build --build-arg VERSION="$(VERSION_TAG)" --rm -t $(DOCKER_IMAGE_NAME):$(VERSION_TAG) .

ca-roots: ## Generate the list of valid CA certificates
	@docker run -dit --rm --name ca-roots debian:stable-slim
	@docker exec --privileged ca-roots sh -c "apt update"
	@docker exec --privileged ca-roots sh -c "apt install -y ca-certificates"
	@docker exec --privileged ca-roots sh -c "cat /etc/ssl/certs/* > /ca-roots.crt"
	@docker cp ca-roots:/ca-roots.crt ca-roots.crt
	@docker stop ca-roots

test: ## Run all tests excluding the vendor dependencies
	# Static analysis
	golangci-lint run ./...
	go-consistent -v ./...

	# Unit tests
	go test -race -cover -v -failfast ./...
