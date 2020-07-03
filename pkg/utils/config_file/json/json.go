package json

import (
	"xip/utils/config_file"
)

type ConfigFileJson struct {
	*config_file.ConfigFile
}

func New(path string) (*ConfigFileJson, error) {
	c, err := config_file.New(path, "json")
	if err != nil {
		return nil, err
	}

	return &ConfigFileJson{c}, nil
}
