package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/buildpack/libbuildpack"
)

func FindExecutableJar(appDir string) ([]libbuildpack.Process, error) {
	if jars, err := filepath.Glob(filepath.Join(appDir, "target", "*.jar")); err == nil {
		for _, jar := range jars {
			// if the Jar has a Main class
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

									webProcess := libbuildpack.Process{
										Type:    "web",
										Command: command,
									}

									return []libbuildpack.Process{webProcess}, nil
								}
							}
						}
						defer fileReader.Close()
					}
				}
			}
		}
	}
	return nil, errors.New("could not file a Jar file")
}
