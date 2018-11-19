package maven

import (
	"path/filepath"
	"bytes"
	"os"
	"os/exec"
	"io"
	"os/user"
	"fmt"
	"errors"

	"github.com/buildpack/libbuildpack"
)

type Runner struct {
	In       []byte
	Out, Err io.Writer
	Command  string
	Options  []string
}

func (r *Runner) Run(appDir, goals string, cache libbuildpack.Cache) (error) {
	err := r.Init(appDir, cache)
	if err != nil {
		return err
	}

	m2Dir, err := r.createMavenRepoDir(appDir, cache)
	if err != nil {
		return err
	}
	defer r.removeMavenRepoSymlink(m2Dir)

	// TODO check MAVEN_CUSTOM_GOALS
	mavenArgs := append(r.Options, goals)

	cmd := exec.Command(r.Command, mavenArgs...)
	cmd.Env = os.Environ()
	cmd.Dir = appDir
	cmd.Stdin = bytes.NewBuffer(r.In)
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err

	if err := cmd.Run(); err != nil {
		return errors.New("failed to run maven command")
	}

	return nil
}

func (r *Runner) Init(appDir string, cache libbuildpack.Cache) (error) {
	mvn, err := r.resolveMavenCommand(appDir, cache)
	if err != nil {
		return err
	}

	r.Command = mvn
	r.Options, err = r.constructMavenOpts(appDir)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runner) resolveMavenCommand(appDir string, cache libbuildpack.Cache) (string, error) {
	if r.hasMavenWrapper(appDir) {
		mvn := filepath.Join(appDir, "mvnw")
		os.Chmod(mvn, 0774)
		return mvn, nil
	} else {
		mavenCacheLayer := cache.Layer("maven")
		mvn, err := r.installMaven(mavenCacheLayer.Root)
		if err != nil {
			return "", err
		}
		return mvn, nil
	}
}

func (r *Runner) installMaven(installDir string) (string, error) {
	// TODO
	return "", nil
}

func (r *Runner) constructMavenOpts(appDir string) ([]string, error) {
	opts := []string{
		"-B",
		"-DskipTests",
		"-DoutputFile=target/mvn-dependency-list.log",
	}

	settingsOpt, err := r.constructMavenSettingsOpts(appDir)
	if err != nil {
		return []string{}, err
	}

	opts = append(opts, settingsOpt)

	// TODO check MAVEN_CUSTOM_OPTS

	return opts, nil
}

func (r *Runner) constructMavenSettingsOpts(appDir string) (string, error) {
	if mvnSettingsPath, isSet := os.LookupEnv("MAVEN_SETTINGS_PATH"); isSet {
		return fmt.Sprintf("-s %s", mvnSettingsPath), nil
	} else if mvnSettingsUrl, isSet := os.LookupEnv("MAVEN_SETTINGS_URL"); isSet {
		settingsXml := filepath.Join(os.TempDir(), "settings.xml")

		cmd := exec.Command("curl", "--retry", "3", "--max-time", "10", "-sfL", mvnSettingsUrl, "-o", settingsXml)
		cmd.Env = os.Environ()
		cmd.Stdout = r.Out
		cmd.Stderr = r.Err

		if err := cmd.Run(); err != nil {
			return "", errors.New("failed to download settings.xml from URL")
		}
		if _, err := os.Stat(settingsXml); os.IsNotExist(err) {
			return "", errors.New(fmt.Sprintf("failed to create %s from URL", settingsXml))
		}
		return fmt.Sprintf("-s %s", settingsXml), nil
	} else if _, err := os.Stat(filepath.Join(appDir, "settings.xml")); !os.IsNotExist(err) {
		return fmt.Sprintf("-s %s", filepath.Join(appDir, "settings.xml")), nil
	}
	return "", nil
}

func (r *Runner) createMavenRepoDir(appDir string, cache libbuildpack.Cache) (string, error) {
	m2Dir, err := defaultMavenHome()
	if err != nil {
		return "", errors.New("error getting maven home")
	}

	m2CacheLayer := cache.Layer("maven_m2")

	err = os.MkdirAll(m2CacheLayer.Root, os.ModePerm)
	if err != nil {
		return "", errors.New("error creating maven cache layer")
	}

	return m2Dir, os.Symlink(m2CacheLayer.Root, m2Dir)
}

func (r *Runner) removeMavenRepoSymlink(m2Dir string) (error) {
	fi, err := os.Lstat(m2Dir)
	if err != nil {
		return err
	}
	if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
		return os.Remove(m2Dir)
	}
	return nil
}

func (r *Runner) hasMavenWrapper(appDir string) (bool) {
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
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(usr.HomeDir, ".m2"), nil
}