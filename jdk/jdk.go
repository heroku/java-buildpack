package jdk

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/heroku/java-buildpack/util"
)

type Installer struct {
	In           []byte
	Out, Err     io.Writer
	Version      Version
	BuildpackDir string
	Log          logger.Logger
}

type Jvm struct {
	Version Version `toml:"version"`
	Home    string  `toml:"home"`
}

type Version struct {
	// Major should be an int but https://github.com/go-yaml/yaml/issues/430
	Major  string  `toml:"major"`
	Tag    string  `toml:"tag"`
	Vendor string  `toml:"vendor"`
}

const (
	DefaultJdkMajorVersion = "8"
	DefaultVendor          = "openjdk"
	DefaultJdkBaseUrl      = "https://lang-jvm.s3.amazonaws.com/jdk"
)

var (
	DefaultVersionStrings = map[string]string{
		"8":  "1.8.0_212",
		"9":  "9.0.4",
		"10": "10.0.2",
		"11": "11.0.3",
		"12": "12.0.1",
	}
)

func (i *Installer) Init(appDir string) error {
	v, err := i.detectVersion(appDir)
	if err != nil {
		return err
	}

	i.Version = v

	return nil
}

func (i *Installer) Install(appDir string, layersDir layers.Layers) (Jvm, error) {
	err := i.Init(appDir)
	if err != nil {
		return Jvm{}, err
	}

	// TODO check the build plan to see if another JDK has already been installed?

	jdkUrl, err := GetVersionUrl(i.Version)
	if err != nil {
		return Jvm{}, err
	}

	if !IsValidJdkUrl(jdkUrl) {
		return Jvm{}, invalidJdkVersion(i.Version.Tag, jdkUrl)
	}

	jdkLayer := layersDir.Layer("jdk")
	jdk := Jvm{
		Home:    jdkLayer.Root,
		Version: i.Version,
	}

	// check to see if there is an existing cache layer with the same Version.Tag as the one we need to install.
	// if that layer exists, we can reuse it and skip this whole business of installing the JDK
	if _, err := os.Stat(jdkLayer.Metadata); err == nil {
		var oldJdkMetadata Jvm
		if err = jdkLayer.ReadMetadata(&oldJdkMetadata); err == nil {
			if oldJdkMetadata.Version.Tag == jdk.Version.Tag {
				i.Log.Info("JDK %s installed from cache", oldJdkMetadata.Version.Tag)
				return oldJdkMetadata, nil
			} else {
				i.Log.Debug("removing expired JDK from cache")
				if err = i.removeLayer(jdkLayer); err != nil {
					return jdk, err
				}
			}
		} else {
			i.Log.Debug(err.Error())
		}
	} else {
		i.Log.Debug("no cached JDK detected")
	}

	if err := i.fetchJdk(jdkUrl, jdkLayer); err != nil {
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

	if err := i.applyJdkOverlay(jdkLayer, appDir); err != nil {
		return jdk, err
	}

	i.Log.Info("JDK %s installed", jdk.Version.Tag)

	jreDir := filepath.Join(jdkLayer.Root, "jre")
	jreLayer := layersDir.Layer("jre")
	if err = i.removeLayer(jreLayer); err != nil {
		return jdk, err
	}

	if _, err = os.Stat(jreDir); err != nil || os.IsNotExist(err) {
		// jdk 11+
		if err := jdkLayer.WriteMetadata(jdk, layers.Launch, layers.Cache, layers.Build); err != nil {
			return jdk, err
		}
	} else {
		if err := i.extractJreFromJdk(jreDir, jreLayer); err != nil {
			return jdk, err
		}

		jre := Jvm{
			Home:    jreLayer.Root,
			Version: i.Version,
		}
		if err := jdkLayer.WriteMetadata(jdk, layers.Cache, layers.Build); err != nil {
			return jdk, err
		}
		if err := jreLayer.WriteMetadata(jre, layers.Launch); err != nil {
			return jre, err
		}
		i.Log.Info("JRE %s added to launch image", jre.Version.Tag)
	}

	return jdk, nil
}

func (i *Installer) fetchJdk(jdkUrl string, layer layers.Layer) error {
	cmd := exec.Command(filepath.Join("jdk-fetcher"), jdkUrl, layer.Root)
	cmd.Env = os.Environ()
	cmd.Stdout = i.Out
	cmd.Stderr = i.Err

	return cmd.Run()
}

func (i *Installer) removeLayer(layer layers.Layer) error {
	if _, err := os.Stat(layer.Metadata); err == nil {
		if err := os.Remove(layer.Metadata); err != nil {
			return err
		}
	}

	cmd := exec.Command(filepath.Join("rm"), "-rf", layer.Root)
	cmd.Env = os.Environ()
	cmd.Stdout = i.Out
	cmd.Stderr = i.Err

	return cmd.Run()
}

func (i *Installer) extractJreFromJdk(jreDir string, jreLayer layers.Layer) error {
	cmd := exec.Command(filepath.Join("cp"), "-R", jreDir, jreLayer.Root)
	cmd.Env = os.Environ()
	cmd.Stdout = i.Out
	cmd.Stderr = i.Err

	return cmd.Run()
}

func (i *Installer) applyJdkOverlay(layer layers.Layer, appDir string) error {
	cmd := exec.Command(filepath.Join("jdk-overlay"), layer.Root, filepath.Join(appDir, ".jdk-overlay"))
	cmd.Env = os.Environ()
	cmd.Stdout = i.Out
	cmd.Stderr = i.Err

	return cmd.Run()
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

func InstallCerts(jdk Jvm) error {
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
			if err = os.Remove(jdkCacerts); err != nil {
				return err
			}
			return os.Symlink(systemCacerts, jdkCacerts)
		}
	}
	return nil
}

func CreateProfileScripts(buildpackDir string, layer layers.Layer) error {
	jvmProfiled, err := ioutil.ReadFile(filepath.Join(buildpackDir, "profile.d", "jvm.sh"))
	if err != nil {
		return err
	}
	if err = layer.WriteProfile("jvm.sh", string(jvmProfiled)); err != nil {
		return err
	}

	jdbcProfiled, err := ioutil.ReadFile(filepath.Join(buildpackDir, "profile.d", "jdbc.sh"))
	if err != nil {
		return err
	}
	if err = layer.WriteProfile("jdbc.sh", string(jdbcProfiled)); err != nil {
		return err
	}

	return nil
}

func defaultVersion() Version {
	version, _ := ParseVersionString(DefaultVersionStrings[DefaultJdkMajorVersion])
	return version
}

func ParseVersionString(v string) (Version, error) {
	if v == "10" || v == "11" {
		return ParseVersionString(DefaultVersionStrings[v])
	} else if m := regexp.MustCompile("^(1[0-9])\\.").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major := m[0][1]
		return Version{
			Vendor: DefaultVendor,
			Tag:    v,
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^1\\.([7-9])$").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major := m[0][1]
		return Version{
			Vendor: DefaultVendor,
			Tag:    DefaultVersionStrings[major],
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^([7-9])$").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major := m[0][1]
		return Version{
			Vendor: DefaultVendor,
			Tag:    DefaultVersionStrings[major],
			Major:  major,
		}, nil
	} else if m := regexp.MustCompile("^1\\.([7-9])").FindAllStringSubmatch(v, -1); len(m) == 1 {
		major := m[0][1]
		return Version{
			Vendor: DefaultVendor,
			Tag:    v,
			Major:  major,
		}, nil
	} else if v == "9+181" || v == "9.0.0" {
		return Version{
			Vendor: DefaultVendor,
			Tag:    "9-181",
			Major:  "9",
		}, nil
	} else if m := regexp.MustCompile("^9\\.").FindAllStringSubmatch(v, -1); len(m) == 1 {
		return Version{
			Vendor: DefaultVendor,
			Tag:    v,
			Major:  "9",
		}, nil
	} else if m := regexp.MustCompile("^zulu-(.*)").FindAllStringSubmatch(v, -1); len(m) == 1 {
		return Version{
			Vendor: "zulu-",
			Tag:    m[0][1],
			Major:  parseMajorVersion(m[0][1]),
		}, nil
	} else if m := regexp.MustCompile("^openjdk-(.*)").FindAllStringSubmatch(v, -1); len(m) == 1 {
		return Version{
			Vendor: "openjdk",
			Tag:    m[0][1],
			Major:  parseMajorVersion(m[0][1]),
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

	vendor := v.Vendor
	if v.Vendor == "zulu" {
		vendor = "zulu-"
	}

	return fmt.Sprintf("%s/%s/%s%s.tar.gz", baseUrl, stack, vendor, v.Tag), nil
}

func IsValidJdkUrl(url string) bool {
	res, err := http.Head(url)
	if err != nil {
		return false
	}
	return res.StatusCode < 300
}

func parseMajorVersion(tag string) string {
	if m := regexp.MustCompile("^1\\.7").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "7"
	} else if m := regexp.MustCompile("^1\\.8").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "8"
	} else if m := regexp.MustCompile("^1\\.9").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "9"
	} else if m := regexp.MustCompile("^7").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "7"
	} else if m := regexp.MustCompile("^8").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "8"
	} else if m := regexp.MustCompile("^9").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "9"
	} else if m := regexp.MustCompile("^10").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "10"
	} else if m := regexp.MustCompile("^11").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "11"
	} else if m := regexp.MustCompile("^12").FindAllStringSubmatch(tag, -1); len(m) == 1 {
		return "12"
	} else {
		return tag
	}
}
