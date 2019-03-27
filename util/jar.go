package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/buildpack/libbuildpack/layers"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func FindExecutableJar(appDir string) (layers.Processes, error) {
	if jars, err := filepath.Glob(filepath.Join(appDir, "target", "*.[jw]ar")); err == nil {
		for _, jar := range jars {
			return detectMainClass(jar)
		}
	}

	return nil, errors.New("could not find a Jar file")
}

func detectMainClass(jar string) (layers.Processes, error) {
	reader, err := zip.OpenReader(jar)
	if err != nil {
		return nil, errors.New("unable to open Jar file")
	} else {
		for _, file := range reader.File {
			if file.Name == "META-INF/MANIFEST.MF" {
				fileReader, err := file.Open()
				if err == nil {
					bytes, err := ioutil.ReadAll(fileReader)
					if err != nil {
						return nil, errors.New("unable to read Jar file")
					} else {
						if strings.Contains(string(bytes), "Main-Class") {
							command := "java"
							if strings.Contains(string(bytes), "Start-Class") {
								command = fmt.Sprintf("%s -Dserver.port=$PORT", command)
							}
							command = fmt.Sprintf("%s -jar target/%s", command, filepath.Base(jar))

							webProcess := layers.Process{
								Type:    "web",
								Command: command,
							}

							return layers.Processes{webProcess}, nil
						}
					}
				}
				defer fileReader.Close()
			}
		}
	}
	return layers.Processes{}, nil
}