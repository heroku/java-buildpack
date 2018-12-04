package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/procfile"
	"github.com/heroku/java-buildpack/util"
)

var (
	launchDir string
)

func init() {
	cmd.FlagLaunch(&launchDir)
}

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	cmd.Exit(writeLaunchMetadata(launchDir))
}

func writeLaunchMetadata(launchDir string) error {
	appDir, err := os.Getwd()
	if err != nil {
		return err
	}

	logger := libbuildpack.NewLogger(ioutil.Discard, os.Stdout)

	processes, err := procfile.Parse(filepath.Join(appDir, "Procfile"))
	if err != nil {
		logger.Debug(err.Error())
	} else {
		logProcessTypes(processes, logger)
		return writeMetadata(launchDir, processes, logger)
	}

	processes, err = util.FindExecutableJar(appDir)
	if err != nil {
		logger.Debug(err.Error())
	} else {
		logProcessTypes(processes, logger)
		return writeMetadata(launchDir, processes, logger)
	}

	logger.Info("No process types detected")
	return nil
}

func writeMetadata(launchDir string, processes []libbuildpack.Process, logger libbuildpack.Logger) error {
	launch := libbuildpack.Launch{
		Root:   launchDir,
		Logger: logger,
	}

	return launch.WriteMetadata(libbuildpack.LaunchMetadata{
		Processes: processes,
	})
}

func logProcessTypes(processes []libbuildpack.Process, logger libbuildpack.Logger) {
	logger.Info("Discovered process type(s):")
	for _, p := range processes {
		logger.Info("  %s: %s\n", p.Type, p.Command)
	}
}
