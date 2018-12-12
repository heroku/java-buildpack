// +build integration

package maven_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/heroku/java-buildpack/jdk"
	"github.com/heroku/java-buildpack/maven"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
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
		layersDir      layers.Layers
		stdout, stderr *bytes.Buffer
	)

	it.Before(func() {
		os.Setenv("STACK", "heroku-18")
		wd, _ := os.Getwd()
		os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), filepath.Join(wd, "..", "bin")))

		layersDir = layers.NewLayers(os.TempDir(), logger.DefaultLogger())

		stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
		runner = &maven.Runner{
			In:  []byte{},
			Out: io.MultiWriter(stdout, it.Out()),
			Err: io.MultiWriter(stderr, it.Out()),
		}
	})

	it.After(func() {
		os.RemoveAll(layersDir.Root)
	})

	when("#Install", func() {
		it("should get the maven version", func() {
			err := runner.Run(fixture("app_with_pom"), "--version", []string{}, layersDir)
			if err != nil {
				t.Fatal(stderr.String(), err)
			}

			expected := "Apache Maven"
			if !strings.Contains(stdout.String(), expected) {
				t.Fatalf("Expected to find \"%s\" in: %s", expected, stdout)
			}
		})

		it("should run maven", func() {
			err := runner.Run(fixture("app_with_pom"), "clean install", []string{}, layersDir)
			if err != nil {
				t.Fatal(stderr.String(), err)
			}

			expected := "[INFO] BUILD SUCCESS"
			if !strings.Contains(stdout.String(), expected) {
				t.Fatalf("Expected to find \"%s\" in: %s", expected, stdout)
			}
		})

		it("should use settings.xml", func() {
			err := runner.Run(fixture("app_with_settings"), "clean install", []string{}, layersDir)
			if err != nil {
				t.Fatal(stderr.String(), err)
			}

			expected := "[INFO] Downloading from jboss-public-repository: http://repository.jboss.org/nexus/content/groups/public"
			if !strings.Contains(stdout.String(), expected) {
				t.Fatalf("Expected to find \"%s\" in: %s", expected, stdout)
			}

			expected = "[INFO] BUILD SUCCESS"
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

				if err := runner.Init(appDir, layersDir); err != nil {
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

	layersRoot, err := ioutil.TempDir("", "jdk-cache")
	if err != nil {
		return err
	}
	jdkLayers := layers.NewLayers(layersRoot, logger.DefaultLogger())

	jdkInstaller := jdk.Installer{
		In:           []byte{},
		Out:          os.Stdout,
		Err:          os.Stderr,
		BuildpackDir: filepath.Join(wd, ".."),
	}
	jdkInstall, err := jdkInstaller.Install("/", jdkLayers)
	if err != nil {
		return err
	}

	os.Setenv("JAVA_HOME", jdkInstall.Home)
	os.Setenv("PATH", fmt.Sprintf("%s/bin:%s", os.Getenv("PATH"), jdkInstall.Home))

	return nil
}
