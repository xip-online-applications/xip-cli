package config

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"gopkg.in/ini.v1"
)

type Config struct {
	FileName string

	SsoEntries   map[string]ConfigEntrySso
	AliasEntries map[string]ConfigEntryAlias
}

type ConfigEntry struct {
	Name string `ini:"-"`

	Region string `ini:"region"`
	Output string `ini:"output"`
}

type ConfigEntrySso struct {
	Name string `ini:"-"`

	Region    string `ini:"region"`
	Output    string `ini:"output"`
	StartUrl  string `ini:"sso_start_url"`
	AccountId string `ini:"sso_account_id"`
	Role      string `ini:"sso_role_name"`
	SsoRegion string `ini:"sso_region"`
}

type ConfigEntryAlias struct {
	Name string `ini:"-"`

	Region        string `ini:"region"`
	Output        string `ini:"output"`
	SourceProfile string `ini:"source_profile"`
	RoleArn       string `ini:"role_arn"`
}

func NewConfig() Config {
	usr, _ := user.Current()

	return Config{
		FileName: usr.HomeDir + "/.aws/config",

		SsoEntries:   map[string]ConfigEntrySso{},
		AliasEntries: map[string]ConfigEntryAlias{},
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

		if section.HasKey("sso_start_url") {
			configEntry := ConfigEntrySso{}
			if err := section.MapTo(&configEntry); err != nil {
				return fmt.Errorf("could not parse sso config entry %s", sectionName)
			}

			configEntry.Name = profileName
			c.SsoEntries[configEntry.Name] = configEntry
		} else {
			configEntry := ConfigEntryAlias{}
			if err := section.MapTo(&configEntry); err != nil {
				return fmt.Errorf("could not parse alias config entry %s", sectionName)
			}

			configEntry.Name = profileName
			c.AliasEntries[configEntry.Name] = configEntry
		}
	}

	return nil
}

func (c *Config) Save() error {
	file := ini.Empty()

	for _, configEntry := range c.SsoEntries {
		section, _ := file.NewSection(c.getSectionName(configEntry.Name))
		if err := section.ReflectFrom(&configEntry); err != nil {
			panic(err)
		}
	}

	for _, configEntry := range c.AliasEntries {
		section, _ := file.NewSection(c.getSectionName(configEntry.Name))
		if err := section.ReflectFrom(&configEntry); err != nil {
			panic(err)
		}
	}

	if _, err := os.Stat(c.FileName); err != nil {
		_ = os.MkdirAll(path.Dir(c.FileName), 0777)
	}

	err := file.SaveTo(c.FileName)
	if err != nil {
		return fmt.Errorf("could not save the config file to %s: %s", c.FileName, err.Error())
	}

	return nil
}

func (c *Config) GetProfile(Profile string) (ConfigEntry, error) {
	if val, ok := c.AliasEntries[Profile]; ok {
		return ConfigEntry{
			Name:   val.Name,
			Region: val.Region,
			Output: val.Output,
		}, nil
	}

	if val, ok := c.SsoEntries[Profile]; ok {
		return ConfigEntry{
			Name:   val.Name,
			Region: val.Region,
			Output: val.Output,
		}, nil
	}

	return ConfigEntry{}, fmt.Errorf("config entry not found for profile %s", Profile)
}

func (c *Config) GetAliasProfile(Profile string) (ConfigEntryAlias, error) {
	if val, ok := c.AliasEntries[Profile]; ok {
		return val, nil
	}

	return ConfigEntryAlias{}, fmt.Errorf("alias config entry not found for profile %s", Profile)
}

func (c *Config) GetSsoProfile(Profile string) (ConfigEntrySso, error) {
	if val, ok := c.SsoEntries[Profile]; ok {
		return val, nil
	}

	return ConfigEntrySso{}, fmt.Errorf("sso config entry not found for profile %s", Profile)
}

func (c *Config) SetSsoProfile(Profile string, Region string, Output string, StartUrl string, AccountId string, Role string, SsoRegion string) error {
	ssoEntry := ConfigEntrySso{
		Name:      Profile,
		Region:    Region,
		Output:    Output,
		StartUrl:  StartUrl,
		AccountId: AccountId,
		Role:      Role,
		SsoRegion: SsoRegion,
	}

	c.SsoEntries[ssoEntry.Name] = ssoEntry
	return c.Save()
}

func (c *Config) SetAliasProfile(Profile string, Region string, Output string, SourceProfile string, RoleArn string) error {
	aliasEntry := ConfigEntryAlias{
		Name:          Profile,
		Region:        Region,
		Output:        Output,
		SourceProfile: SourceProfile,
		RoleArn:       RoleArn,
	}

	c.AliasEntries[aliasEntry.Name] = aliasEntry
	return c.Save()
}

func (c *Config) getSectionName(name string) string {
	if name == "default" {
		return name
	}

	return "profile " + name
}
