.PHONY: all release clean build lint generate help

# VERSION=`git describe --tags 2>/dev/null || echo ""`
# BUILDTIME=`date +%FT%T%z`
# BUILDHASH=`git rev-parse --short HEAD`
# BUILDDIRTY=`[ $$(git status --short | wc -c) -ne 0 ] && echo "-dirty"`

# LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.buildtime=${BUILDTIME} -X main.buildhash=${BUILDHASH}${BUILDDIRTY}"

all: generate lint test build ## Test, lint check and build application

release: clean ## Build release version of application
	mkdir -p ./dist
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -a -o ./dist/http_fetcher_exporter_darvin_amd64
	#  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ./dist/http_fetcher_exporter_linux_amd64

clean:
	rm -rf ./dist/*

build: ## Build application
	go build .

lint: ## Lint the project
	golangci-lint run ./...

generate: ## Generate mocks
	go generate ./...


test:
	go test -race ./...

help: ## Print all possible targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {gsub("\\\\n",sprintf("\n%22c",""), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)