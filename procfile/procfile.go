package procfile

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/buildpack/libbuildpack/layers"
	"gopkg.in/yaml.v2"
)

func Parse(file string) (layers.Processes, error) {
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
				processes := layers.Processes{}
				for name, command := range processTypes {
					processes = append(processes, layers.Process{
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
