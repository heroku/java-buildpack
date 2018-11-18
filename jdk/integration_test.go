// +build integration

package jdk_test

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
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
		wd, _ := os.Getwd()

		os.Setenv("STACK", "heroku-18")

		cacerts, err := ioutil.ReadFile(filepath.Join(wd, "..", "test", "fixtures", "cacerts"));
		if err != nil {
			t.Fatal(err)
		}
		cacertsFile := filepath.Clean("/etc/ssl/certs/java/cacerts")
		err = os.MkdirAll(filepath.Dir(cacertsFile), 0755)
		if err != nil {
			t.Fatal(err)
		}
		fh, err := os.OpenFile(cacertsFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer fh.Close()
		_, err = io.Copy(fh, strings.NewReader(string(cacerts)))
		if err != nil {
			t.Fatal(err)
		}

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

			// TODO also check that it's a symlink
			if _, err := os.Stat(filepath.Join(launch.Layer("jdk").Root, "jre", "lib", "security", "cacerts")); os.IsNotExist(err) {
				t.Fatal("cacerts not linked")
			}

			if _, err := os.Stat(filepath.Join(launch.Layer("jdk").Root, "profile.d", "jvm.sh")); os.IsNotExist(err) {
				t.Fatal("JVM profile.d script not installed")
			}

			if _, err := os.Stat(filepath.Join(launch.Layer("jdk").Root, "profile.d", "jdbc.sh")); os.IsNotExist(err) {
				t.Fatal("JDBC profile.d script not installed")
			}

			var jdkMetadata jdk.Jdk
			if err := launch.Layer("jdk").ReadMetadata(&jdkMetadata); err != nil {
				t.Fatal("Layer metadata was not written")
			}

			if jdkMetadata.Home != launch.Layer("jdk").Root {
				t.Fatalf(`Jdk.Home did not match: got %s, want %s`, jdkMetadata.Home, launch.Layer("jdk").Root)
			}

			if jdkMetadata.Version.Major != 8 {
				t.Fatalf(`Jdk.Version.Tag did not match: got %s, want %s`, jdkMetadata.Version.Major, 8)
			}

			if jdkMetadata.Version.Tag != jdk.DefaultVersionStrings[8] {
				t.Fatalf(`Jdk.Version.Tag did not match: got %s, want %s`, jdkMetadata.Version.Tag, jdk.DefaultVersionStrings[8])
			}

			if jdkMetadata.Version.Vendor != jdk.DefaultVersionStrings[8] {
				t.Fatalf(`Jdk.Version.Vendor did not match: got %s, want %s`, jdkMetadata.Version.Vendor, jdk.DefaultVendor)
			}
		})
	})
}

