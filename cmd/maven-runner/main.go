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
	goals       string
	platformDir string
	cacheDir    string
	launchDir   string
)

func init() {
	cmd.FlagGoals(&goals)
	cmd.FlagPlatform(&platformDir)
	cmd.FlagCache(&cacheDir)
	cmd.FlagLaunch(&launchDir)
}

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	cmd.Exit(runGoals(goals, platformDir, cacheDir, launchDir))
}

func runGoals(goals, platformDir, cacheDir, launchDir string) (error) {

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

	print("---> Installing JDK")
	jdkInstaller := jdk.Installer{
		In:  []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
	}
	jdkInstall, err := jdkInstaller.Install(appDir, cache, launch)
	if err != nil {
		return err
	}

	// FIXME
	// ideally the jdk pkg would do this, but it's hard to undo. even more preferably, the jdk stuff would be in
	// it's own buildpack, and the lifecycle would handle this.
	os.Setenv("JAVA_HOME", jdkInstall.Home)
	os.Setenv("PATH", fmt.Sprintf("%s/bin:%s", os.Getenv("PATH"), jdkInstall.Home))

	print("---> Running Maven")
	runner := maven.Runner{
		In:  []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
	}
	err = runner.Run(appDir, goals, cache)
	if err != nil {
		return err
	}

	// TODO write launch.toml

	return nil
}
