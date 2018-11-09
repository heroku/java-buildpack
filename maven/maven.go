package maven

import (
	"path/filepath"
	"bytes"
	"os"
	"os/exec"
	"io"

	"github.com/buildpack/libbuildpack"
	"os/user"
	"fmt"
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

	// TODO check MAVEN_CUSTOM_GOALS
	mavenArgs := append(r.Options, goals)

	cmd := exec.Command(r.Command, mavenArgs...)
	cmd.Env = os.Environ()
	cmd.Dir = appDir
	cmd.Stdin = bytes.NewBuffer(r.In)
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err

	if err := cmd.Run(); err != nil {
		return err
	}

	err = r.removeMavenRepoSymlink(m2Dir)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runner) Init(appDir string, cache libbuildpack.Cache) (error) {
	mvn, err := r.resolveMavenCommand(appDir, cache)
	if err != nil {
		return err
	}

	r.Command = mvn
	r.Options = r.constructMavenOpts(appDir)
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

func (r *Runner) constructMavenOpts(appDir string) ([]string) {
	opts := []string{
		"-B",
		"-DskipTests",
		"-DoutputFile=target/mvn-dependency-list.log",
	}

	opts = append(opts, r.constructMavenSettingsOpts(appDir))

	// TODO check MAVEN_CUSTOM_OPTS

	return opts
}

func (r *Runner) constructMavenSettingsOpts(appDir string) (string) {
	if _, isSet := os.LookupEnv("MAVEN_SETTINGS_PATH"); isSet {
		// TODO
	} else if _, isSet := os.LookupEnv("MAVEN_SETTINGS_URL"); isSet {
		// TODO
	} else if _, err := os.Stat(filepath.Join(appDir, "settings.xml")); !os.IsNotExist(err) {
		return fmt.Sprintf("-s %s", filepath.Join(appDir, "settings.xml"))
	}
	return ""
}

func (r *Runner) createMavenRepoDir(appDir string, cache libbuildpack.Cache) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	m2Dir := filepath.Join(usr.HomeDir, ".m2")
	m2CacheLayer := cache.Layer("maven_m2")

	err = os.MkdirAll(m2CacheLayer.Root, os.ModePerm)
	if err != nil {
		return "", err
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
