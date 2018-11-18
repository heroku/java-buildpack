package jdk

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"os"
	"regexp"
	"strconv"
	"errors"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/heroku/java-buildpack/util"
	"github.com/buildpack/libbuildpack"
)

type Installer struct {
	In           []byte
	Out, Err     io.Writer
	Version      Version
	BuildpackDir string
}

type Jdk struct {
	Version Version `toml:"version"`
	Home    string  `toml:"home"`
}

type Version struct {
	Major  int    `toml:"major"`
	Tag    string `toml:"tag"`
	Vendor string `toml:"vendor"`
}

const (
	DefaultJdkMajorVersion = 8
	DefaultVendor          = "openjdk"
	DefaultJdkBaseUrl      = "https://lang-jvm.s3.amazonaws.com/jdk"
)

var (
	DefaultVersionStrings = map[int]string{
		8:  "1.8.0_191",
		9:  "9.0.4",
		10: "10.0.2",
		11: "11.0.1",
	}
)

func (i *Installer) Init(appDir string) (error) {
	v, err := i.detectVersion(appDir)
	if err != nil {
		return err
	}

	i.Version = v

	return nil
}

func (i *Installer) Install(appDir string, cache libbuildpack.Cache, launchDir libbuildpack.Launch) (Jdk, error) {
	i.Init(appDir)
	// check the build plan to see if another JDK has already been installed?

	jdkUrl, err := GetVersionUrl(i.Version)
	if err != nil {
		return Jdk{}, err
	}

	if !IsValidJdkUrl(jdkUrl) {
		return Jdk{}, errors.New(fmt.Sprintf("Invalid JDK version: %s", jdkUrl))
	}

	jdkLayer := launchDir.Layer("jdk")
	jdk := Jdk{
		Home:    jdkLayer.Root,
		Version: i.Version,
	}

	cmd := exec.Command(filepath.Join("jdk-fetcher"), jdkUrl, jdkLayer.Root)
	cmd.Env = os.Environ()
	cmd.Stdout = i.Out
	cmd.Stderr = i.Err

	if err := cmd.Run(); err != nil {
		return jdk, err
	}

	if err := InstallCerts(jdk); err != nil {
		return jdk, err
	}

	if err := CreateProfileScripts(i.BuildpackDir, jdkLayer); err != nil {
		return jdk, err
	}

	// TODO install pgconfig
	// TODO install metrics agent
	// TODO apply the .jdk-overlay

	if err := jdk.WriteMetadata(jdkLayer); err != nil {
		return jdk, err
	}

	return jdk, nil
}

func (jdk Jdk) WriteMetadata(layer libbuildpack.LaunchLayer) error {
	return layer.WriteMetadata(jdk)
}

func (i *Installer) detectVersion(appDir string) (Version, error) {
	systemPropertiesFile := filepath.Join(appDir, "system.properties")
	if _, err := os.Stat(systemPropertiesFile); !os.IsNotExist(err) {
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

func InstallCerts(jdk Jdk) error {
	jreCacerts := filepath.Join(jdk.Home, "jre", "lib", "security", "cacerts")
	jdkCacerts := filepath.Join(jdk.Home, "lib", "security", "cacerts")
	systemCacerts := filepath.Clean("/etc/ssl/certs/java/cacerts")

	if _, err := os.Stat(systemCacerts); !os.IsNotExist(err) {
		if _, err := os.Stat(jreCacerts); !os.IsNotExist(err) {
			if err = os.Remove(jreCacerts); err != nil {
				return err
			}
			return os.Symlink(systemCacerts, jreCacerts)
		} else if _, err := os.Stat(jdkCacerts); !os.IsNotExist(err) {
			if err = os.Remove(jreCacerts); err != nil {
				return err
			}
			return os.Symlink(systemCacerts, jdkCacerts)
		}
	}
	return nil
}

func CreateProfileScripts(buildpackDir string, layer libbuildpack.LaunchLayer) error {
	jvmProfiled, err := ioutil.ReadFile(filepath.Join(buildpackDir, "profile.d", "jvm.sh"));
	if err != nil {
		return err
	}
	layer.WriteProfile("jvm.sh", string(jvmProfiled))

	jdbcProfiled, err := ioutil.ReadFile(filepath.Join(buildpackDir, "profile.d", "jdbc.sh"));
	if err != nil {
		return err
	}
	layer.WriteProfile("jdbc.sh", string(jdbcProfiled))

	return nil
}

func defaultVersion() Version {
	version, _ := ParseVersionString(DefaultVersionStrings[DefaultJdkMajorVersion])
	return version
}

func ParseVersionString(v string) (Version, error) {
	if v == "10" || v == "11" {
		major, _ := strconv.Atoi(v)
		return ParseVersionString(DefaultVersionStrings[major])
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

func GetVersionUrl(v Version) (string, error) {
	baseUrl := DefaultJdkBaseUrl
	if customBaseUrl, ok := os.LookupEnv("DEFAULT_JDK_BASE_URL"); ok {
		baseUrl = customBaseUrl
	}

	stack, ok := os.LookupEnv("STACK")
	if !ok {
		return "", errors.New("missing stack")
	}

	return fmt.Sprintf("%s/%s/%s%s.tar.gz", baseUrl, stack, v.Vendor, v.Tag), nil
}

func IsValidJdkUrl(url string) bool {
	res, err := http.Head(url)
	if err != nil {
		return false
	}
	return res.StatusCode < 300
}
