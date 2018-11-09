// +build integration

package jdk_test

import (
	"fmt"
	"os"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/heroku/java-buildpack/jdk"
	"github.com/buildpack/libbuildpack"
)

func TestIntegrationJdk(t *testing.T) {
	spec.Run(t, "Installer", testIntegrationJdk, spec.Report(report.Terminal{}))
}

func testIntegrationJdk(t *testing.T, when spec.G, it spec.S) {
	var (
		installer *jdk.Installer
		cache     libbuildpack.Cache
		launch    libbuildpack.Launch
	)

	it.Before(func() {
		os.Setenv("STACK", "heroku-18")

		wd, _ := os.Getwd()
		os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), filepath.Join(wd, "..", "bin")))

		installer = &jdk.Installer{
			In:  []byte{},
			Out: os.Stdout,
			Err: os.Stderr,
		}

		logger := libbuildpack.NewLogger(ioutil.Discard, ioutil.Discard)

		cacheRoot, err := ioutil.TempDir("", "cache")
		if err != nil {
			t.Fatal(err)
		}

		launchRoot, err := ioutil.TempDir("", "launch")
		if err != nil {
			t.Fatal(err)
		}

		cache = libbuildpack.Cache{Root: cacheRoot, Logger: logger}
		launch = libbuildpack.Launch{Root: launchRoot, Logger: logger}
	})

	it.After(func() {
		os.RemoveAll(cache.Root)
		os.RemoveAll(launch.Root)
	})

	when("#Install", func() {
		it("should install the default", func() {
			_, err := installer.Install(fixture("app_with_jdk_version"), cache, launch)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := os.Stat(launch.Layer("jdk").Root); os.IsNotExist(err) {
				t.Fatal("launch layer not created")
			}

			if _, err := os.Stat(filepath.Join(launch.Layer("jdk").Root, "bin", "java")); os.IsNotExist(err) {
				t.Fatal("java not installed")
			}
		})
	})
}

