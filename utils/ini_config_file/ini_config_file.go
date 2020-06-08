package ini_config_file

import (
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type ConfigFileIni struct {
	path  string
	viper *viper.Viper
}

func New(path string) (*ConfigFileIni, error) {
	c := new(ConfigFileIni)
	c.path = path

	if _, err := os.Stat(c.path); err != nil {
		_ = os.MkdirAll(filepath.Dir(c.path), os.ModePerm)
		_ = ioutil.WriteFile(c.path, []byte(""), 0644)
	}

	c.viper = viper.New()
	c.viper.SetConfigName(filepath.Base(c.path))
	c.viper.AddConfigPath(filepath.Dir(c.path))
	c.viper.SetConfigType("ini")

	if err := c.Read(); err != nil {
		return nil, fmt.Errorf("Fatal error reading config file: %s \n", err)
	}

	return c, nil
}

func AppConf() *ConfigFileIni {
	usr, _ := user.Current()
	conf, err := New(usr.HomeDir + "/.xip/config")

	if err != nil {
		panic(fmt.Errorf("Could not open the configuration file: %s \n", err))
	}

	return conf
}

func (c *ConfigFileIni) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

func (c *ConfigFileIni) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *ConfigFileIni) Set(key string, value interface{}) {
	c.viper.Set(key, value)
}

func (c *ConfigFileIni) Read() error {
	if err := c.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error reading config file: %s \n", err)
	}

	return nil
}

func (c *ConfigFileIni) Write() error {
	if err := c.viper.WriteConfigAs(c.path); err != nil {
		return fmt.Errorf("Fatal error writing config file: %s \n", err)
	}

	return nil
}
