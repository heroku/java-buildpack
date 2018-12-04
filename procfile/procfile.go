package procfile

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/buildpack/libbuildpack"
	"gopkg.in/yaml.v2"
)

func Parse(file string) ([]libbuildpack.Process, error) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, errors.New("failed to read Procfile")
		} else {
			var processTypes map[string]string
			err := yaml.Unmarshal(data, &processTypes)
			if err != nil {
				return nil, errors.New("failed to parse Procfile")
			} else {
				processes := []libbuildpack.Process{}
				for name, command := range processTypes {
					processes = append(processes, libbuildpack.Process{
						Type:    name,
						Command: command,
					})
				}

				return processes, nil
			}
		}
	}
	return nil, errors.New("could not find Procfile")
}
