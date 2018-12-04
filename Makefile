.EXPORT_ALL_VARIABLES:

.PHONY: test \
        build \
        clean \
        package \
        release

SHELL=/bin/bash -o pipefail

GO111MODULE := on

VERSION := "v$$(cat buildpack.toml | grep version | sed -e 's/version = //g' | xargs)"

test:
	-docker rm -f java-buildpack-test
	@docker create --name java-buildpack-test --workdir /app golang:1.11 bash -c "go test ./... -tags=integration"
	@docker cp . java-buildpack-test:/app
	@docker start -a java-buildpack-test

build:
	@GOOS=linux go build -o "bin/jdk-installer" ./cmd/jdk-installer/...
	@GOOS=linux go build -o "bin/maven-runner" ./cmd/maven-runner/...
	@GOOS=linux go build -o "bin/releaser" ./cmd/releaser/...

clean:
	-rm -f java-buildpack-$(VERSION).tgz

package: clean build
	@tar cvzf java-buildpack-$(VERSION).tgz bin/ profile.d/ buildpack.toml README.md LICENSE

release:
	@git tag $(VERSION)
	@git push --tags origin master