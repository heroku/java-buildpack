package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/procfile"
	"github.com/heroku/java-buildpack/util"
)

var (
	layersDir string
)

func init() {
	cmd.FlagLayers(&layersDir)
}

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	cmd.Exit(writeLaunchMetadata(layersDir))
}

func writeLaunchMetadata(launchDir string) error {
	appDir, err := os.Getwd()
	if err != nil {
		return err
	}

	log := logger.DefaultLogger()

	processes, err := procfile.Parse(filepath.Join(appDir, "Procfile"))
	if err != nil {
		log.Debug(err.Error())
	} else {
		logProcessTypes(processes, log)
		return writeMetadata(launchDir, processes, log)
	}

	processes, err = util.FindExecutableJar(appDir)
	if err != nil {
		log.Debug(err.Error())
	} else {
		logProcessTypes(processes, log)
		return writeMetadata(launchDir, processes, log)
	}

	log.Info("No process types detected")
	return nil
}

func writeMetadata(layersRoot string, processes layers.Processes, log logger.Logger) error {
	layersDir := layers.NewLayers(layersRoot, log)

	return layersDir.WriteMetadata(layers.Metadata{
		Processes: processes,
	})
}

func logProcessTypes(processes layers.Processes, log logger.Logger) {
	log.Info("Discovered process type(s):")
	for _, p := range processes {
		log.Info("  %s: %s\n", p.Type, p.Command)
	}
}
