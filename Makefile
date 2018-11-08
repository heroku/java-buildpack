SHELL=/bin/bash -o pipefail

GO111MODULE := on

.PHONY: test \
		build

test:
	@go test ./...

build:
	@go build -o "bin/maven-runner" ./cmd/maven-runner/...