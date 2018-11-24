package main

import (
	"path/filepath"
	"flag"
	"os"

	"github.com/buildpack/libbuildpack"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/procfile"
	"github.com/heroku/java-buildpack/util"
)

var (
	launchDir   string
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

func writeLaunchMetadata(launchDir string) (error) {
	println("\n[Releasing]")

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}

	logger := libbuildpack.NewLogger(os.Stdout, os.Stdout)

	processes, err := procfile.Parse(filepath.Join(appDir, "Procfile"))
	if err != nil {
		logger.Debug(err.Error())
	} else {
		return writeMetadata(launchDir, processes, logger)
	}

	processes, err = util.FindExecutableJar(appDir)
	if err != nil {
		logger.Debug(err.Error())
	} else {
		return writeMetadata(launchDir, processes, logger)
	}

	logger.Debug("no process types detected")
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
