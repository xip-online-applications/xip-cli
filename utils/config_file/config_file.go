package config_file

import (
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ConfigFile struct {
	configType string
	path       string
	viper      *viper.Viper
}

func New(path string, configType string) (*ConfigFile, error) {
	c := new(ConfigFile)
	c.path = path
	c.configType = configType

	if _, err := os.Stat(c.path); err != nil {
		_ = os.MkdirAll(filepath.Dir(c.path), os.ModePerm)
		_ = ioutil.WriteFile(c.path, []byte(""), 0644)
	}

	c.viper = viper.New()
	c.viper.SetConfigName(filepath.Base(c.path))
	c.viper.AddConfigPath(filepath.Dir(c.path))
	c.viper.SetConfigType(configType)

	if err := c.Read(); err != nil {
		return nil, fmt.Errorf("Fatal error reading config file: %s \n", err)
	}

	return c, nil
}

func (c *ConfigFile) SetConfigType(configType string) {
	c.configType = configType
}

func (c *ConfigFile) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

func (c *ConfigFile) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *ConfigFile) GetStringOptional(key string) *string {
	if !c.IsSet(key) {
		return nil
	}

	val := c.viper.GetString(key)

	return &val
}

func (c *ConfigFile) Set(key string, value *string) {
	if value == nil {
		c.viper.Set(key, "")
	} else {
		c.viper.Set(key, *value)
	}
}

func (c *ConfigFile) Keys() []string {
	return c.viper.AllKeys()
}

func (c *ConfigFile) Read() error {
	if err := c.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error reading config file: %s \n", err)
	}

	return nil
}

func (c *ConfigFile) Write() error {
	if err := c.viper.WriteConfigAs(c.path); err != nil {
		return fmt.Errorf("Fatal error writing config file: %s \n", err)
	}

	return nil
}
