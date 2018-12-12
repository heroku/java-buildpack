package procfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/heroku/java-buildpack/procfile"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestProcfile(t *testing.T) {
	spec.Run(t, "Procfile", testProcfile, spec.Report(report.Terminal{}))
}

func testProcfile(t *testing.T, when spec.G, it spec.S) {
	when("#Parse", func() {
		it("should find process types", func() {
			processes, err := procfile.Parse(filepath.Join(fixture("app_with_procfile"), "Procfile"))

			if err != nil {
				t.Fatal(err)
			}

			if len(processes) != 2 {
				t.Fatalf(`Did not find process types: got %d, want %d`, len(processes), 2)
			}

			if found, p := findProcessType(processes, "web"); found {
				expected := "java -cp target/classes:target/dependency/* com.example.Main"
				if p.Command != expected {
					t.Fatalf(`Did not find correct command: got %s, want %s`, p.Command, expected)
				}
			} else {
				t.Fatal("Did not find a web process")
			}
		})
	})
}

func findProcessType(processes layers.Processes, name string) (bool, layers.Process) {
	for _, p := range processes {
		if p.Type == name {
			return true, p
		}
	}
	return false, layers.Process{}
}

func fixture(name string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "test", "fixtures", name)
}
