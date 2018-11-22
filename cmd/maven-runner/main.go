package main

import (
	"flag"
	"io/ioutil"
	"os"
	"fmt"

	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/maven"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/jdk"
)

var (
	goals        string
	options      string
	platformDir  string
	cacheDir     string
	launchDir    string
	buildpackDir string
)

func init() {
	flag.StringVar(&goals, "goals", "clean install", "maven goals to run")
	flag.StringVar(&options, "goals", "", "maven goals to run")

	cmd.FlagPlatform(&platformDir)
	cmd.FlagCache(&cacheDir)
	cmd.FlagLaunch(&launchDir)

	// TODO shouldn't we be able to find this from the binary?
	cmd.FlagBuildpack(&buildpackDir)
}

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	cmd.Exit(runGoals(goals, options, platformDir, cacheDir, launchDir, buildpackDir))
}

func runGoals(goals, options, platformDir, cacheDir, launchDir, buildpackDir string) (error) {
	logger := libbuildpack.NewLogger(ioutil.Discard, ioutil.Discard)

	platform, err := libbuildpack.NewPlatform(platformDir, logger)
	if err != nil {
		return err
	}

	platform.Envs.SetAll()

	cache := libbuildpack.Cache{Root: cacheDir, Logger: logger}
	launch := libbuildpack.Launch{Root: launchDir, Logger: logger}

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}

	println("\n[Installing JDK]")
	jdkInstaller := jdk.Installer{
		In:           []byte{},
		Out:          os.Stdout,
		Err:          os.Stderr,
		BuildpackDir: buildpackDir,
	}
	jdkInstall, err := jdkInstaller.Install(appDir, cache, launch)
	if err != nil {
		return err
	}
	println("Java", jdkInstall.Version.Tag, "installed")

	// FIXME
	// ideally the jdk pkg would do this, but it's hard to undo. even more preferably, the jdk stuff would be in
	// it's own buildpack, and the lifecycle would handle this.
	os.Setenv("JAVA_HOME", jdkInstall.Home)
	os.Setenv("PATH", fmt.Sprintf("%s/bin:%s", os.Getenv("PATH"), jdkInstall.Home))

	println("\n[Running Maven]")
	runner := maven.Runner{
		In:  []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
	}

	if err = runner.Run(appDir, goals, []string{options}, cache); err != nil {
		return err
	}

	return nil
}
