package main

import (
	"github.com/buildpack/libbuildpack"
	"io/ioutil"
	"github.com/heroku/java-buildpack/maven"
	"os"
	"github.com/heroku/java-buildpack/cmd"
	"flag"
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

	// TODO install jdk

	runner := maven.Runner{
		In: []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
	}

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}
	runner.Run(appDir, goals, cache)

	// TODO write launch.toml

	return nil
}
