// +build integration

package maven_test

import (
	"path/filepath"
	"bytes"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"testing"
	"strings"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/jdk"
	"github.com/heroku/java-buildpack/maven"
)

func TestIntegrationMaven(t *testing.T) {
	jdkRoot, err := ioutil.TempDir("", "jdk-root")
	if err != nil {
		t.Fatal(err)
	}

	err = installGlobalJdk(jdkRoot)
	if err != nil {
		t.Fatal(err)
	}

	spec.Run(t, "Runner", testIntegrationMaven, spec.Report(report.Terminal{}))

	os.RemoveAll(jdkRoot)
}

func testIntegrationMaven(t *testing.T, when spec.G, it spec.S) {
	var (
		runner         *maven.Runner
		cache          libbuildpack.Cache
		launch         libbuildpack.Launch
		stdout, stderr *bytes.Buffer
	)

	it.Before(func() {
		os.Setenv("STACK", "heroku-18")
		wd, _ := os.Getwd()
		os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), filepath.Join(wd, "..", "bin")))

		logger := libbuildpack.NewLogger(ioutil.Discard, ioutil.Discard)

		cacheRoot, err := ioutil.TempDir("", "cache")
		if err != nil {
			t.Fatal(err)
		}
		cache = libbuildpack.Cache{Root: cacheRoot, Logger: logger}

		//launchRoot, err := ioutil.TempDir("", "launch")
		//if err != nil {
		//	t.Fatal(err)
		//}
		//launch = libbuildpack.Launch{Root: launchRoot, Logger: logger}
		//
		//jdkInstaller := jdk.Installer{
		//	In:  []byte{},
		//	Out: os.Stdout,
		//	Err: os.Stderr,
		//}
		//jdkInstall, err := jdkInstaller.Install(fixture("app_with_pom"), cache, launch)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//
		//os.Setenv("JAVA_HOME", jdkInstall.Home)
		//os.Setenv("PATH", fmt.Sprintf("%s/bin:%s", os.Getenv("PATH"), jdkInstall.Home))

		stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
		runner = &maven.Runner{
			In:  []byte{},
			Out: io.MultiWriter(stdout, it.Out()),
			Err: io.MultiWriter(stderr, it.Out()),
		}
	})

	it.After(func() {
		os.RemoveAll(cache.Root)
		os.RemoveAll(launch.Root)
	})

	when("#Install", func() {
		it("should get the maven version", func() {
			err := runner.Run(fixture("app_with_pom"), "--version", cache)
			if err != nil {
				t.Fatal(stderr.String(), err)
			}

			expected := "Apache Maven"
			if !strings.Contains(stdout.String(), expected) {
				t.Fatalf("Expected to find \"%s\" in: %s", expected, stdout)
			}
		})

		it("should run maven", func() {
			err := runner.Run(fixture("app_with_pom"), "clean install", cache)
			if err != nil {
				t.Fatal(stderr.String(), err)
			}

			expected := "[INFO] BUILD SUCCESS"
			if !strings.Contains(stdout.String(), expected) {
				t.Fatalf("Expected to find \"%s\" in: %s", expected, stdout)
			}
		})
	})

	when("#Init", func() {
		when("MAVEN_SETTINGS_URL is set", func() {
			it("should not use the default settings", func() {
				appDir := fixture("app_with_settings")
				expected := "https://raw.githubusercontent.com/kissaten/settings-xml-example/master/settings.xml"
				os.Setenv("MAVEN_SETTINGS_URL", expected)

				if err := runner.Init(appDir, cache); err != nil {
					t.Fatal(err)
				}

				if !hasOption(runner.Options, "-s /tmp/settings.xml") {
					t.Fatalf(`runner options does not use environment variable: \n%s`, runner.Options)
				}
			})

			it.After(func() {
				os.Unsetenv("MAVEN_SETTINGS_URL")
			})
		})
	})
}

func installGlobalJdk(installDir string) error {
	os.Setenv("STACK", "heroku-18")
	wd, _ := os.Getwd()
	os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), filepath.Join(wd, "..", "bin")))

	logger := libbuildpack.NewLogger(ioutil.Discard, ioutil.Discard)

	cacheRoot, err := ioutil.TempDir("", "jdk-cache")
	if err != nil {
		return err
	}
	jdkCache := libbuildpack.Cache{Root: cacheRoot, Logger: logger}

	jdkLaunch := libbuildpack.Launch{Root: installDir, Logger: logger}

	jdkInstaller := jdk.Installer{
		In:  []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
		BuildpackDir: filepath.Join(wd, ".."),
	}
	jdkInstall, err := jdkInstaller.Install("/", jdkCache, jdkLaunch)
	if err != nil {
		return err
	}

	os.Setenv("JAVA_HOME", jdkInstall.Home)
	os.Setenv("PATH", fmt.Sprintf("%s/bin:%s", os.Getenv("PATH"), jdkInstall.Home))

	return nil
}