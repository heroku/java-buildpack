package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/maven"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/jdk"
)

type MavenEnv interface {
	AddRootDir(baseDir string) error
	AddEnvDir(envDir string) error
	List() []string
}

var (
	goals string
)

func init() {
	cmd.FlagGoals(&goals)
}

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "not enough arguments"))
	}

	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	platformDir := args[0]
	cacheDir := args[1]
	launchDir := args[2]

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

	jdkInstaller := jdk.Installer{
		In: []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
	}
	jdkInstaller.Install(appDir, cache, launch)

	runner := maven.Runner{
		In: []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
	}
	runner.Run(appDir, goals, cache)

	// TODO write launch.toml

	return nil
}
