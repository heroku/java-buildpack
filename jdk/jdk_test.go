package jdk_test

import (
	"testing"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/heroku/java-buildpack/jdk"
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