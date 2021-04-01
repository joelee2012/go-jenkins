
help: ## Print usage
	@sed -r '/^(\w+):[^#]*##/!d;s/^([^:]+):[^#]*##\s*(.*)/\x1b[36m\1\t:\x1b[m \2/g' ${MAKEFILE_LIST}

REPO_ROOT:=$(shell git rev-parse --show-toplevel)

build: ## build package
	go build -v .

test: ## run test
	cat $(REPO_ROOT)/env.sh && \
	. $(REPO_ROOT)/env.sh && \
	env && \
	go test -coverprofile=coverage.out -v &&\
	go tool cover -html=coverage.out -o cover.html

fmt: ## format code
	go fmt .

lint: ## lint code
	go get honnef.co/go/tools/cmd/staticcheck && staticcheck .

.PHONY: test build


