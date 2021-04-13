package config

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Config struct {
	FileName       string
	DefaultProfile *EntryConfig

	ConfigEntries map[string]EntryConfig
}

func NewConfig() Config {
	usr, _ := user.Current()

	return Config{
		FileName: filepath.FromSlash(usr.HomeDir + "/.aws/config"),

		ConfigEntries: map[string]EntryConfig{},
	}
}

func LoadConfig() (Config, error) {
	configFile := NewConfig()
	if err := configFile.Load(); err != nil {
		return Config{}, fmt.Errorf("culd not load the config file: %s", err.Error())
	}

	return configFile, nil
}

func (c *Config) Load() error {
	file, err := ini.Load(c.FileName)
	if err != nil {
		file = ini.Empty()
	}

	for _, sectionName := range file.SectionStrings() {
		if sectionName == "DEFAULT" {
			continue
		}

		section := file.Section(sectionName)
		profileName := sectionName

		if sectionName != "default" {
			profileName = sectionName[8:]
		}

		configEntry, err := EntryFromSection(profileName, section)
		if err != nil {
			return err
		}
		c.ConfigEntries[configEntry.Name] = *configEntry

		if configEntry.IsDefaultProfile() {
			c.DefaultProfile = configEntry
		}
	}

	return nil
}

func (c *Config) Save() error {
	file, err := ini.Load(c.FileName)
	if err != nil {
		file = ini.Empty()
	}

	if c.DefaultProfile != nil {
		name := c.getSectionName("default")
		section, err := file.GetSection(name)
		if err != nil {
			section, _ = file.NewSection(name)
		}

		if err := section.ReflectFrom(&c.DefaultProfile); err != nil {
			panic(err)
		}
	}

	for _, configEntry := range c.ConfigEntries {
		if configEntry.Name == "default" {
			continue
		}

		name := c.getSectionName(configEntry.Name)
		section, err := file.GetSection(name)
		if err != nil {
			section, _ = file.NewSection(name)
		}

		if err := section.ReflectFrom(&configEntry); err != nil {
			panic(err)
		}
	}

	if _, err := os.Stat(c.FileName); err != nil {
		_ = os.MkdirAll(path.Dir(c.FileName), 0777)
	}

	err = file.SaveTo(c.FileName)
	if err != nil {
		return fmt.Errorf("could not save the config file to %s: %s", c.FileName, err.Error())
	}

	return nil
}

func (c *Config) GetProfile(Profile string) (EntryConfig, error) {
	if val, ok := c.ConfigEntries[Profile]; ok {
		return val, nil
	}

	return EntryConfig{}, fmt.Errorf("config entry not found for profile %s", Profile)
}

func (c *Config) GetAliasProfile(Profile string) (EntryConfig, error) {
	if val, ok := c.GetProfile(Profile); ok == nil && val.IsAliasProfile() {
		return val, nil
	}

	return EntryConfig{}, fmt.Errorf("config entry not found for profile %s", Profile)
}

func (c *Config) GetSsoProfile(Profile string) (EntryConfig, error) {
	if val, ok := c.GetProfile(Profile); ok == nil && val.IsSsoProfile() {
		return val, nil
	}

	return EntryConfig{}, fmt.Errorf("config entry not found for profile %s", Profile)
}

func (c *Config) SetSsoProfile(Profile string, Region string, Output string, StartUrl string, AccountId string, Role string, SsoRegion string) error {
	ssoEntry := EntryConfig{
		Name:      Profile,
		Region:    Region,
		Output:    Output,
		StartUrl:  StartUrl,
		AccountId: AccountId,
		Role:      Role,
		SsoRegion: SsoRegion,
	}

	c.ConfigEntries[ssoEntry.Name] = ssoEntry
	return c.Save()
}

func (c *Config) SetAliasProfile(Profile string, Region string, Output string, SourceProfile string, RoleArn string) error {
	aliasEntry := EntryConfig{
		Name:          Profile,
		Region:        Region,
		Output:        Output,
		SourceProfile: SourceProfile,
		RoleArn:       RoleArn,
	}

	c.ConfigEntries[aliasEntry.Name] = aliasEntry
	return c.Save()
}

func (c *Config) SetDefaultProfile(Profile string) (EntryConfig, error) {
	AliasProfile, err := c.GetProfile(Profile)
	if err != nil {
		return EntryConfig{}, err
	}

	DefaultProfile := AliasProfile
	DefaultProfile.SetIsDefaultProfile()

	c.DefaultProfile = &DefaultProfile

	return AliasProfile, nil
}

func (c *Config) getSectionName(name string) string {
	if name == "default" {
		return name
	}

	return "profile " + name
}
