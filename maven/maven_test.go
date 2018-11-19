package maven_test

import (
	"io/ioutil"
	"strings"
	"path/filepath"
	"fmt"
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/heroku/java-buildpack/maven"
	"github.com/buildpack/libbuildpack"
)

func TestMaven(t *testing.T) {
	spec.Run(t, "Runner", testMaven, spec.Report(report.Terminal{}))
}

func testMaven(t *testing.T, when spec.G, it spec.S) {
	var (
		runner *maven.Runner
		cache  libbuildpack.Cache
		appDir string
	)

	it.Before(func() {
		cache = libbuildpack.Cache{
			Root:   os.TempDir(),
			Logger: libbuildpack.NewLogger(ioutil.Discard, ioutil.Discard),
		}

		runner = &maven.Runner{
			In:  []byte{},
			Out: os.Stdout,
			Err: os.Stderr,
		}
	})

	it.After(func() {
		os.RemoveAll(cache.Root)
	})

	when("#Init", func() {
		when("has a maven wrapper", func() {
			appDir = fixture("app_with_wrapper")

			it("should use the mvnw command", func() {
				if err := runner.Init(appDir, cache); err != nil {
					t.Fatal(err)
				}

				if !strings.HasSuffix(runner.Command, "mvnw") {
					t.Fatalf(`runner command does not use wrapper: \n%s`, runner.Command)
				}
			})
		})
		when("has a settings file", func() {
			appDir = fixture("app_with_settings")

			it("should use the settings option", func() {
				if err := runner.Init(appDir, cache); err != nil {
					t.Fatal(err)
				}

				if !hasOption(runner.Options, fmt.Sprintf("-s %s/settings.xml", appDir)) {
					t.Fatalf(`runner options does not use settings.xml: \n%s`, runner.Options)
				}
			})
		})
		when("MAVEN_SETTINGS_PATH is set", func() {
			appDir = fixture("app_with_settings")

			it("should not use the default settings", func() {
				expected := "any/old/path/settings.xml"
				os.Setenv("MAVEN_SETTINGS_PATH", expected)

				if err := runner.Init(appDir, cache); err != nil {
					t.Fatal(err)
				}

				if !hasOption(runner.Options, fmt.Sprintf("-s %s", expected)) {
					t.Fatalf(`runner options does not use environment variable: \n%s`, runner.Options)
				}
			})

			it.After(func() {
				os.Unsetenv("MAVEN_SETTINGS_PATH")
			})
		})
	})
}

func hasOption(opts []string, opt string) (bool) {
	for _, b := range opts {
		if b == opt {
			return true
		}
	}
	return false
}

func fixture(name string) (string) {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "test", "fixtures", name)
}
