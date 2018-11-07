package jdk_test

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/heroku/java-buildpack/jdk"
	"github.com/google/go-cmp/cmp"
)

func TestJdk(t *testing.T) {
	spec.Run(t, "Installer", testJdk, spec.Report(report.Terminal{}))
}

func testJdk(t *testing.T, when spec.G, it spec.S) {
	var (
	//installer *jdk.Installer
	)

	it.Before(func() {
		// TODO
		os.Setenv("STACK", "heroku-18")
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
				t.Fatalf(`syscall.Exec Argv did not match: (-got +want)\n%s`, diff)
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
				t.Fatalf(`syscall.Exec Argv did not match: (-got +want)\n%s`, diff)
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

		it("should not parse garbage", func() {
			expected := "1bh"
			_, err := jdk.ParseVersionString(expected)
			if err == nil {
				t.Fatal("unexpected success")
			}
		})
	})
}
