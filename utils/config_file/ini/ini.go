package ini

import (
	"fmt"
	"os/user"
	"xip/utils/config_file"
)

type ConfigFileIni struct {
	*config_file.ConfigFile
}

func New(path string) (*ConfigFileIni, error) {
	c, err := config_file.New(path, "ini")
	if err != nil {
		return nil, err
	}

	return &ConfigFileIni{c}, nil
}

func AppConf() *ConfigFileIni {
	usr, _ := user.Current()
	conf, err := New(usr.HomeDir + "/.xip/config")

	if err != nil {
		panic(fmt.Errorf("Could not open the configuration file: %s \n", err))
	}

	return conf
}
