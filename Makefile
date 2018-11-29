.EXPORT_ALL_VARIABLES:

SHELL=/bin/bash -o pipefail

GO111MODULE := on

VERSION := $$(cat buildpack.toml | grep version | sed -e 's/version = //g' | xargs)

test:
	-docker rm -f java-buildpack-test
	@docker create --name java-buildpack-test --workdir /app golang:1.11 bash -c "go test ./... -tags=integration"
	@docker cp . java-buildpack-test:/app
	@docker start -a java-buildpack-test

build:
	@GOOS=linux go build -o "bin/maven-runner" ./cmd/maven-runner/...
	@GOOS=linux go build -o "bin/releaser" ./cmd/releaser/...

clean:
	-rm -f java-buildpack-v$(VERSION).tgz

release: clean build
	@tar cvzf java-buildpack-v$(VERSION).tgz bin/ profile.d/ buildpack.toml README.md LICENSE