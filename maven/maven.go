package maven

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/buildpack/libbuildpack/layers"
)

type Runner struct {
	In       []byte
	Out, Err io.Writer
	Command  string
	Options  []string
	Goals    []string
}

func (r *Runner) Run(appDir, defaultGoals string, options []string, layersDir layers.Layers) error {
	r.Goals = parseGoals(defaultGoals)
	r.Options = trimArgs(options)

	err := r.Init(appDir, layersDir)
	if err != nil {
		return err
	}

	m2Dir, err := r.createMavenRepoDir(appDir, layersDir)
	if err != nil {
		return err
	}
	defer r.removeMavenRepoSymlink(m2Dir)

	mavenArgs := append(r.Options, r.Goals...)

	fmt.Printf("$ mvn %s %s\n", strings.Join(r.Options, " "), strings.Join(r.Goals, " "))
	cmd := exec.Command(r.Command, mavenArgs...)
	cmd.Env = os.Environ()
	cmd.Dir = appDir
	cmd.Stdin = bytes.NewBuffer(r.In)
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err

	if err := cmd.Run(); err != nil {
		return failedToRunMaven(err)
	}

	return nil
}

// This function should remain free of side-effects to the filesystem
func (r *Runner) Init(appDir string, layersDir layers.Layers) error {
	mvn, err := r.resolveMavenCommand(appDir, layersDir)
	if err != nil {
		return err
	}

	r.Command = mvn
	r.Options, err = r.constructOptions(appDir)
	if err != nil {
		return err
	}

	r.Goals = r.constructGoals(r.Goals)

	return nil
}

func (r *Runner) resolveMavenCommand(appDir string, layersDir layers.Layers) (string, error) {
	if r.hasMavenWrapper(appDir) {
		mvn := filepath.Join(appDir, "mvnw")
		os.Chmod(mvn, 0774)
		return mvn, nil
	} else {
		mavenCacheLayer := layersDir.Layer("maven")
		mvn, err := r.installMaven(mavenCacheLayer.Root)
		if err != nil {
			return "", err
		}
		return mvn, nil
	}
}

func (r *Runner) installMaven(installDir string) (string, error) {
	cmd := exec.Command(filepath.Join("maven-installer"), installDir)
	cmd.Env = os.Environ()
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err

	return filepath.Join(installDir, "bin", "mvn"), cmd.Run()
}

func (r *Runner) constructGoals(defaultGoals []string) []string {
	if goals, isSet := os.LookupEnv("MAVEN_CUSTOM_GOALS"); isSet {
		return parseGoals(goals)
	}
	return defaultGoals
}

func (r *Runner) constructOptions(appDir string) ([]string, error) {
	opts := []string{
		"-B",
		"-DoutputFile=target/dependencies.txt",
	}

	opts = append(opts, r.Options...)

	settingsOpt, err := r.constructSettingsOpts(appDir)
	if err != nil {
		return []string{}, err
	}

	opts = append(opts, settingsOpt...)

	if customOpts, isSet := os.LookupEnv("MAVEN_CUSTOM_OPTS"); isSet {
		opts = append(opts, parseGoals(customOpts)...)
	}

	return trimArgs(opts), nil
}

func (r *Runner) constructSettingsOpts(appDir string) ([]string, error) {
	if mvnSettingsPath, isSet := os.LookupEnv("MAVEN_SETTINGS_PATH"); isSet {
		return []string{"-s", mvnSettingsPath}, nil
	} else if mvnSettingsUrl, isSet := os.LookupEnv("MAVEN_SETTINGS_URL"); isSet {
		settingsXml := filepath.Join(os.TempDir(), "settings.xml")

		out, err := os.Create(settingsXml)
		if err != nil {
			return nil, failedToDownloadSettings(err)
		}
		defer out.Close()

		resp, err := http.Get(mvnSettingsUrl)
		if err != nil {
			return nil, failedToDownloadSettings(err)
		}
		defer resp.Body.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return nil, failedToDownloadSettings(err)
		}

		if _, err := os.Stat(settingsXml); os.IsNotExist(err) {
			return nil, failedToDownloadSettingsFromUrl(mvnSettingsUrl, err)
		}
		return []string{"-s", settingsXml}, nil
	} else if _, err := os.Stat(filepath.Join(appDir, "settings.xml")); !os.IsNotExist(err) {
		return []string{"-s", "settings.xml"}, nil
	}
	return nil, nil
}

func (r *Runner) createMavenRepoDir(appDir string, layersDir layers.Layers) (string, error) {
	m2Dir, err := defaultMavenHome()
	if err != nil {
		return "", errors.New(fmt.Sprintf("error getting maven home: %s", err))
	}

	m2CacheLayer := layersDir.Layer("maven_m2")

	err = os.MkdirAll(m2CacheLayer.Root, os.ModePerm)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error creating maven cache layer: %s", err))
	}

	return m2Dir, os.Symlink(m2CacheLayer.Root, m2Dir)
}

func (r *Runner) removeMavenRepoSymlink(m2Dir string) error {
	fi, err := os.Lstat(m2Dir)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		return os.Remove(m2Dir)
	}
	return nil
}

func (r *Runner) hasMavenWrapper(appDir string) bool {
	_, err := os.Stat(filepath.Join(appDir, "mvnw"))
	if !os.IsNotExist(err) {
		_, err = os.Stat(filepath.Join(appDir, ".mvn", "wrapper", "maven-wrapper.jar"))
		if !os.IsNotExist(err) {
			_, err = os.Stat(filepath.Join(appDir, ".mvn", "wrapper", "maven-wrapper.properties"))
			if !os.IsNotExist(err) {
				return true
			}
		}
	}
	return false
}

func defaultMavenHome() (string, error) {
	home, found := os.LookupEnv("HOME")
	if found {
		return filepath.Join(home, ".m2"), nil
	}
	return "", errors.New("could not find user home")
}

func parseGoals(goals string) []string {
	return trimArgs(strings.Split(goals, " "))
}

func trimArgs(args []string) []string {
	var parsedArgs []string
	for _, rawArg := range args {
		arg := strings.TrimSpace(rawArg)
		if arg != "" {
			parsedArgs = append(parsedArgs, arg)
		}
	}
	return parsedArgs
}
