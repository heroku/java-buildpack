package main

import (
	"flag"
	"os"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/buildpack/libbuildpack/platform"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/maven"
)

var (
	goals        string
	options      string
	platformRoot string
	layersRoot   string
)

func init() {
	flag.StringVar(&goals, "goals", "clean install", "maven goals to run")
	flag.StringVar(&options, "options", "", "maven goals to run")

	cmd.FlagPlatform(&platformRoot)
	cmd.FlagLayers(&layersRoot)
}

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	cmd.Exit(runGoals(goals, options, platformRoot, layersRoot))
}

func runGoals(goals, options, platformRoot, layersRoot string) error {
	log := logger.DefaultLogger()

	platformDir, err := platform.DefaultPlatform(platformRoot, log)
	if err != nil {
		return err
	}

	err = platformDir.EnvironmentVariables.SetAll()
	if err != nil {
		return err
	}

	layersDir := layers.NewLayers(layersRoot, log)

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}

	runner := maven.Runner{
		In:  []byte{},
		Out: os.Stdout,
		Err: os.Stderr,
	}

	if err = runner.Run(appDir, goals, []string{options}, layersDir); err != nil {
		return err
	}

	return nil
}
