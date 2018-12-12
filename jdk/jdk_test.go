package jdk_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/google/go-cmp/cmp"
	"github.com/heroku/java-buildpack/jdk"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestJdk(t *testing.T) {
	spec.Run(t, "Installer", testJdkInstaller, spec.Report(report.Terminal{}))
	spec.Run(t, "Jdk", testJdk, spec.Report(report.Terminal{}))
}

func testJdk(t *testing.T, when spec.G, it spec.S) {
	var (
		layersDir layers.Layers
	)

	it.Before(func() {
		root, err := ioutil.TempDir("", "layers")
		if err != nil {
			t.Fatal(err)
		}
		layersDir = layers.NewLayers(root, logger.DefaultLogger())
	})

	it.After(func() {
		os.RemoveAll(layersDir.Root)
	})

	when("#WriteMetadata", func() {
		it("should detect jdk version", func() {
			expected := jdk.Jdk{
				Home: layersDir.Root,
				Version: jdk.Version{
					Major:  8,
					Tag:    "1.8.0_191",
					Vendor: "openjdk",
				},
			}

			if err := expected.WriteMetadata(layersDir.Layer("jdk")); err != nil {
				t.Fatal(err)
			}

			var actual jdk.Jdk
			if err := layersDir.Layer("jdk").ReadMetadata(&actual); err != nil {
				t.Fatal("Layer metadata was not written")
			}

			if actual.Home != expected.Home {
				t.Fatalf(`Jdk.Home did not match: got %s, want %s`, actual.Home, expected.Home)
			}

			if actual.Version.Major != expected.Version.Major {
				t.Fatalf(`Jdk.Version.Tag did not match: got %d, want %d`, actual.Version.Major, expected.Version.Major)
			}

			if actual.Version.Tag != expected.Version.Tag {
				t.Fatalf(`Jdk.Version.Tag did not match: got %s, want %s`, actual.Version.Tag, expected.Version.Tag)
			}

			if actual.Version.Vendor != expected.Version.Vendor {
				t.Fatalf(`Jdk.Version.Vendor did not match: got %s, want %s`, actual.Version.Vendor, expected.Version.Vendor)
			}
		})
	})

}

func testJdkInstaller(t *testing.T, when spec.G, it spec.S) {
	var (
		installer *jdk.Installer
		layersDir layers.Layers
	)

	it.Before(func() {
		os.Setenv("STACK", "heroku-18")

		installer = &jdk.Installer{
			In:  []byte{},
			Out: os.Stdout,
			Err: os.Stderr,
		}

		layersRoot, err := ioutil.TempDir("", "layers")
		if err != nil {
			t.Fatal(err)
		}
		log := logger.DefaultLogger()
		layersDir = layers.NewLayers(layersRoot, log)
	})

	it.After(func() {
		os.RemoveAll(layersDir.Root)
	})

	when("#Init", func() {
		it("should detect jdk version", func() {
			err := installer.Init(fixture("app_with_jdk_version"))
			if err != nil {
				t.Fatal(err)
			}

			expected := "1.8.0_181"
			if installer.Version.Tag != expected {
				t.Fatalf(`JDK version did not match: got %s, want %s`, installer.Version.Tag, expected)
			}
		})
	})

	when("#GetVersionUrl", func() {
		it("should get 10.0.2", func() {
			url, err := jdk.GetVersionUrl(jdk.Version{
				Major:  10,
				Tag:    "10.0.2",
				Vendor: "openjdk",
			})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(url, "https://lang-jvm.s3.amazonaws.com/jdk/heroku-18/openjdk10.0.2.tar.gz"); diff != "" {
				t.Fatalf(`URL did not match: (-got +want)\n%s`, diff)
			}
		})

		it("should get 1.8.0_181", func() {
			url, err := jdk.GetVersionUrl(jdk.Version{
				Major:  8,
				Tag:    "1.8.0_181",
				Vendor: "openjdk",
			})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(url, "https://lang-jvm.s3.amazonaws.com/jdk/heroku-18/openjdk1.8.0_181.tar.gz"); diff != "" {
				t.Fatalf(`URL did not match: (-got +want)\n%s`, diff)
			}
		})

		it("should get zulu-1.8.0_181", func() {
			url, err := jdk.GetVersionUrl(jdk.Version{
				Major:  8,
				Tag:    "1.8.0_181",
				Vendor: "zulu",
			})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(url, "https://lang-jvm.s3.amazonaws.com/jdk/heroku-18/zulu-1.8.0_181.tar.gz"); diff != "" {
				t.Fatalf(`URL did not match: (-got +want)\n%s`, diff)
			}
		})
	})

	when("#ParseVersionString", func() {
		it("should parse 10.0.2", func() {
			expected := "10.0.2"
			v, err := jdk.ParseVersionString(expected)
			if err != nil {
				t.Fatal(err)
			}

			if v.Major != 10 {
				t.Fatalf(`JDK version did not match: got %d, want %d`, v.Major, 10)
			}

			if v.Tag != expected {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Tag, expected)
			}

			if v.Vendor != "openjdk" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Vendor, "openjdk")
			}
		})

		it("should parse 1.8", func() {
			expected := "1.8"
			v, err := jdk.ParseVersionString(expected)
			if err != nil {
				t.Fatal(err)
			}

			if v.Major != 8 {
				t.Fatalf(`JDK version did not match: got %d, want %d`, v.Major, 8)
			}

			if v.Tag != jdk.DefaultVersionStrings[8] {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Tag, jdk.DefaultVersionStrings[8])
			}

			if v.Vendor != "openjdk" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Vendor, "openjdk")
			}
		})

		it("should parse 11", func() {
			expected := "11"
			v, err := jdk.ParseVersionString(expected)
			if err != nil {
				t.Fatal(err)
			}

			if v.Major != 11 {
				t.Fatalf(`JDK version did not match: got %d, want %d`, v.Major, 11)
			}

			if v.Tag != jdk.DefaultVersionStrings[11] {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Tag, jdk.DefaultVersionStrings[11])
			}

			if v.Vendor != "openjdk" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Vendor, "openjdk")
			}
		})

		it("should parse 9", func() {
			expected := "9+181"
			v, err := jdk.ParseVersionString(expected)
			if err != nil {
				t.Fatal(err)
			}

			if v.Major != 9 {
				t.Fatalf(`JDK version did not match: got %d, want %d`, v.Major, 9)
			}

			if v.Tag != "9-181" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Tag, "9-181")
			}

			if v.Vendor != "openjdk" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Vendor, "openjdk")
			}
		})

		it("should parse zulu", func() {
			expected := "zulu-1.8.0_191"
			v, err := jdk.ParseVersionString(expected)
			if err != nil {
				t.Fatal(err)
			}

			if v.Major != 8 {
				t.Fatalf(`JDK version did not match: got %d, want %d`, v.Major, 8)
			}

			if v.Tag != "1.8.0_191" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Tag, "1.8.0_191")
			}

			if v.Vendor != "zulu-" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Vendor, "zulu")
			}
		})

		it("should parse openjdk", func() {
			expected := "openjdk-1.8.0_191"
			v, err := jdk.ParseVersionString(expected)
			if err != nil {
				t.Fatal(err)
			}

			if v.Major != 8 {
				t.Fatalf(`JDK version did not match: got %d, want %d`, v.Major, 8)
			}

			if v.Tag != "1.8.0_191" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Tag, "1.8.0_191")
			}

			if v.Vendor != "openjdk" {
				t.Fatalf(`JDK version did not match: got %s, want %s`, v.Vendor, "openjdk")
			}
		})

		it("should not parse garbage", func() {
			expected := "1bh"
			_, err := jdk.ParseVersionString(expected)
			if err == nil {
				t.Fatal("unexpected success")
			}
		})
	})
}

func fixture(name string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "test", "fixtures", name)
}
