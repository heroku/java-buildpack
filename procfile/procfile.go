package procfile

import (
	"gopkg.in/yaml.v2"
	"github.com/buildpack/libbuildpack"
	"os"
	"io/ioutil"
	"errors"
)

type Procfile struct {
	ProcessTypes map[string]string
}

func Parse(file string) ([]libbuildpack.Process, error) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, errors.New("failed to read Procfile")
		} else {
			var procfileStruct Procfile
			err := yaml.Unmarshal(data, &procfileStruct)
			if err != nil {
				return nil, errors.New("failed to parse Procfile")
			} else {
				processes := []libbuildpack.Process{}
				for name, command := range procfileStruct.ProcessTypes {
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