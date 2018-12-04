package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/maven"
)

var (
	goals       string
	options     string
	platformDir string
	cacheDir    string
)

func init() {
	flag.StringVar(&goals, "goals", "clean install", "maven goals to run")
	flag.StringVar(&options, "options", "", "maven goals to run")

	cmd.FlagPlatform(&platformDir)
	cmd.FlagCache(&cacheDir)
}

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	cmd.Exit(runGoals(goals, options, platformDir, cacheDir))
}

func runGoals(goals, options, platformDir, cacheDir string) error {
	logger := libbuildpack.NewLogger(ioutil.Discard, os.Stdout)

	platform, err := libbuildpack.NewPlatform(platformDir, logger)
	if err != nil {
		return err
	}

	platform.Envs.SetAll()

	cache := libbuildpack.Cache{Root: cacheDir, Logger: logger}

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}

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
