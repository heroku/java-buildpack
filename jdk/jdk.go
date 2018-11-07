package jdk

import (
	"io"
	"github.com/buildpack/libbuildpack"
	"path/filepath"
	"os"
	"github.com/heroku/java-buildpack/util"
	"regexp"
	"strconv"
	"errors"
)

type Installer struct {
	In       []byte
	Out, Err io.Writer
	Version  Version
}

type Version struct {
	Major  int
	Tag    string
	Vendor string
}

const (
	DefaultVendor    = "openjdk"
)

var (
	DefaultVersionStrings = map[int]string{
		8: "1.8.0_181",
		9: "9.0.1",
		10: "10.0.2",
		11: "11.0.1",
	}
)

func (i *Installer) Init(appDir string) (error) {
	// get the version from system.properties

	v, err := i.detectVersion(appDir)
	if err != nil {
		return err
	}

	i.Version = v

	return nil
}

func (i *Installer) Install(appDir string, cache libbuildpack.Cache, launchDir libbuildpack.Launch) (error) {
	i.Init(appDir)
	// check the build plan to see if another JDK has already been installed?

	// install the jdk
	// apply the overlay

	return nil
}

func (i *Installer) detectVersion(appDir string) (Version, error) {
	systemPropertiesFile := filepath.Join(appDir, "system.properties")
	if _, err := os.Stat(systemPropertiesFile); !os.IsNotExist(err) {
		// read it
		sysProps, err := util.ReadPropertiesFile(systemPropertiesFile)
		if err != nil {
			return defaultVersion(), nil
		}

		if version, ok := sysProps["java.runtime.version"]; ok {
			return ParseVersionString(version)
		}
	}
	return defaultVersion(), nil
}

func defaultVersion() Version {
	version, _ := ParseVersionString(DefaultVersionStrings[8])
	return version
}

func ParseVersionString(v string) (Version, error) {
	if v == "10" {
		return ParseVersionString(DefaultVersionStrings[10])
	} else if v == "11" {
		return ParseVersionString(DefaultVersionStrings[11])
	} else if m := regexp.MustCompile("^(1[0-1])\\.").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    v,
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^1\\.([7-9])$").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    DefaultVersionStrings[major],
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^([7-9])$").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    DefaultVersionStrings[major],
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^1\\.([7-9])").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major, _ := strconv.Atoi(m[0][1])
		return Version{
			Vendor: DefaultVendor,
			Tag:    v,
			Major:  major,
		}, nil
	}

	return Version{}, errors.New("unparseable version string")
}
