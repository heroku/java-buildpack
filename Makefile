SHELL=/bin/bash -o pipefail

GO111MODULE := on

.PHONY: test \
		build

test:
	-docker rm -f java-buildpack-test
	@docker create --name java-buildpack-test --workdir /app golang:1.11 bash -c "go test ./... -tags=integration"
	@docker cp . java-buildpack-test:/app
	@docker start -a java-buildpack-test

build:
	@go build -o "bin/maven-runner" ./cmd/maven-runner/...
	@go build -o "bin/releaser" ./cmd/releaser/...

release: build
	@tar cvzf java-buildpack-$$(cat VERSION).tgz bin/ profile.d/ buildpack.toml README.md LICENSE