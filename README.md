# Heroku Cloud Native Buildpack for Java

[![Build
Status](https://travis-ci.com/heroku/java-buildpack.svg?branch=master)](https://travis-ci.com/heroku/java-buildpack) ![GitHub release](https://img.shields.io/github/release/heroku/java-buildpack.svg)

This is a work in progress (WIP) Heroku [Cloud Native Buildpack](https://buildpacks.io/) for Java apps. It uses Maven to build your application and OpenJDK to run it. However, the JDK version can be configured as described below.

## How it works

The buildpack will detect your app as Java if it has a `pom.xml` file, or one of the other POM formats supports by the [Maven Polyglot plugin](https://github.com/takari/polyglot-maven), in its root directory. It will use Maven to execute the build defined by your `pom.xml` and download your dependencies. The `.m2` folder (local maven repository) will be cached between builds for faster dependency resolution, but neither the `mvn` executable or the `.m2` folder will be available in the runtime image.

## Usage

To use this buildpack with [`pack` CLI](https://github.com/buildpack/pack) run the following commands:

```
$ pack build image:tag --builder=heroku/buildpacks
```

## Customizing

This buildpack supports the following environment variables for customization:

* `MAVEN_CUSTOM_GOALS`
* `MAVEN_CUSTOM_OPTS`
* `MAVEN_SETTINGS_PATH`
* `MAVEN_SETTINGS_URL`

You can select the JDK vendor and version using a `system.properties` file as described in the [Heroku documentation on Java](https://devcenter.heroku.com/articles/java-support).

## Development

Run the unit tests (no Internet required):

```
$ go test ./...
```

Run the integration tests (Internet required):

```
$ make test
```

## Licence

MIT
