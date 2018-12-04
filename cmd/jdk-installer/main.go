package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/jdk"
)

var (
	platformDir  string
	cacheDir     string
	launchDir    string
	buildpackDir string
)

func init() {
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

	cmd.Exit(runGoals(platformDir, cacheDir, launchDir, buildpackDir))
}

func runGoals(platformDir, cacheDir, launchDir, buildpackDir string) error {
	logger := libbuildpack.NewLogger(ioutil.Discard, os.Stdout)

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

	return nil
}
