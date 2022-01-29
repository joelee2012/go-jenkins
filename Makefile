
help: ## Print usage
	@sed -r '/^(\w+):[^#]*##/!d;s/^([^:]+):[^#]*##\s*(.*)/\x1b[36m\1\t:\x1b[m \2/g' ${MAKEFILE_LIST}

build: ## build package
	go build -v .

test: ## run test
	env | sort
	go test -race -coverprofile=coverage.out -covermode=atomic -v ./... | tee test.log
	go tool cover -html=coverage.out -o cover.html

fmt: ## format code
	go fmt .

lint: ## lint code
	staticcheck ./...

.PHONY: test build