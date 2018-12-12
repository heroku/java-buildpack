package main

import (
	"flag"
	"os"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/buildpack/libbuildpack/platform"
	"github.com/heroku/java-buildpack/cmd"
	"github.com/heroku/java-buildpack/jdk"
)

var (
	platformRoot  string
	layersRoot    string
	buildpackRoot string
)

func init() {
	cmd.FlagPlatform(&platformRoot)
	cmd.FlagLayers(&layersRoot)

	// TODO shouldn't we be able to find this from the binary?
	cmd.FlagBuildpack(&buildpackRoot)
}

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		cmd.Exit(cmd.FailCode(cmd.CodeInvalidArgs, "parse arguments"))
	}

	cmd.Exit(runGoals(platformRoot, layersRoot, buildpackRoot))
}

func runGoals(platformRoot, layersRoot, buildpackRoot string) error {
	log := logger.DefaultLogger()

	bpPlatform, err := platform.DefaultPlatform(platformRoot, log)
	if err != nil {
		return err
	}

	err = bpPlatform.EnvironmentVariables.SetAll()
	if err != nil {
		return err
	}

	layersDir := layers.NewLayers(layersRoot, log)

	appDir, err := os.Getwd()
	if err != nil {
		return err
	}

	jdkInstaller := jdk.Installer{
		In:           []byte{},
		Out:          os.Stdout,
		Err:          os.Stderr,
		BuildpackDir: buildpackRoot,
	}
	jdkInstall, err := jdkInstaller.Install(appDir, layersDir)
	if err != nil {
		return err
	}
	println("Java", jdkInstall.Version.Tag, "installed")

	return nil
}
