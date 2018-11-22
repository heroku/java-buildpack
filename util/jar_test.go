package util_test

import (
	"os"
	"io/ioutil"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/util"
	"path/filepath"
)

func TestJar(t *testing.T) {
	spec.Run(t, "Jar", testJar, spec.Report(report.Terminal{}))
}

func testJar(t *testing.T, when spec.G, it spec.S) {
	var (
		launch libbuildpack.Launch
	)

	it.Before(func() {
		launchRoot, err := ioutil.TempDir("", "launch")
		if err != nil {
			t.Fatal(err)
		}
		logger := libbuildpack.NewLogger(ioutil.Discard, ioutil.Discard)
		launch = libbuildpack.Launch{Root: launchRoot, Logger: logger}
	})

	it.After(func() {
		os.RemoveAll(launch.Root)
	})

	when("#FindExecutableJar", func() {
		it("should find an executable jar", func() {
			processes, err := util.FindExecutableJar(fixture("app_with_exec_jar"))

			if err != nil {
				t.Fatal(err)
			}

			if len(processes) != 1 {
				t.Fatalf(`Did not find executable JAR: got %d, want %d`, len(processes), 1)
			}

			if processes[0].Type != "web" {
				t.Fatal("Did not create a web process")
			}

			expected := "java -jar target/my-app-1.0-SNAPSHOT-jar-with-dependencies.jar"
			if processes[0].Command != expected {
				t.Fatalf(`Did create correct command: got %s, want %s`, processes[0].Command, expected)
			}
		})
	})
}


func fixture(name string) (string) {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "test", "fixtures", name)
}