package config

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	FileName       string
	DefaultProfile *ConfigEntryAlias

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

	Region          string `ini:"region"`
	Output          string `ini:"output"`
	SourceProfile   string `ini:"source_profile"`
	RoleArn         string `ini:"role_arn"`
	ExternalId      string `ini:"external_id,omitempty"`
	RoleSessionName string `ini:"role_session_name,omitempty"`
}

func NewConfig() Config {
	usr, _ := user.Current()

	return Config{
		FileName: filepath.FromSlash(usr.HomeDir + "/.aws/config"),

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

	defaultProfileName := c.findDefaultProfileName(file)

	for _, sectionName := range file.SectionStrings() {
		if sectionName == "DEFAULT" {
			continue
		}

		section := file.Section(sectionName)
		profileName := sectionName

		if sectionName != "default" {
			profileName = sectionName[8:]
		}

		if c.isSsoProfile(section) {
			configEntry, err := c.buildSsoProfileFromSection(profileName, section)
			if err != nil {
				return err
			}

			c.SsoEntries[configEntry.Name] = *configEntry
		} else if c.isAliasProfile(section) {
			configEntry, err := c.buildAliasProfileFromSection(profileName, section)
			if err != nil {
				return err
			}

			if configEntry.Name != "default" {
				c.AliasEntries[configEntry.Name] = *configEntry
			}

			if defaultProfileName != nil && configEntry.Name == *defaultProfileName {
				c.DefaultProfile = configEntry
			}
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
		name := c.getSectionName(c.DefaultProfile.Name)
		section, err := file.GetSection(name)
		if err != nil {
			section, _ = file.NewSection(name)
		}

		if err := section.ReflectFrom(&c.DefaultProfile); err != nil {
			panic(err)
		}
	}

	for _, configEntry := range c.SsoEntries {
		name := c.getSectionName(configEntry.Name)
		section, err := file.GetSection(name)
		if err != nil {
			section, _ = file.NewSection(name)
		}

		if err := section.ReflectFrom(&configEntry); err != nil {
			panic(err)
		}
	}

	for _, configEntry := range c.AliasEntries {
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

func (c *Config) SetDefaultProfile(Profile string) error {
	AliasProfile, err := c.GetAliasProfile(Profile)
	if err != nil {
		return err
	}

	DefaultProfile := &AliasProfile
	DefaultProfile.RoleSessionName = "xip-cli--" + DefaultProfile.Name
	DefaultProfile.Name = "default"

	c.DefaultProfile = DefaultProfile

	return nil
}

func (c *Config) findDefaultProfileName(file *ini.File) *string {
	for _, sectionName := range file.SectionStrings() {
		if sectionName != "default" {
			continue
		}

		section := file.Section(sectionName)
		if !c.isAliasProfile(section) {
			return nil
		}

		configEntry, err := c.buildAliasProfileFromSection(sectionName, section)
		if err != nil {
			return nil
		}

		if !strings.HasPrefix(configEntry.RoleSessionName, "xip-cli--") {
			return nil
		}

		profileName := configEntry.RoleSessionName[9:]
		return &profileName
	}

	return nil
}

func (c *Config) isSsoProfile(section *ini.Section) bool {
	return section.HasKey("sso_start_url")
}

func (c *Config) isAliasProfile(section *ini.Section) bool {
	return section.HasKey("source_profile") && section.HasKey("role_arn")
}

func (c *Config) getSectionName(name string) string {
	if name == "default" {
		return name
	}

	return "profile " + name
}

func (c *Config) buildSsoProfileFromSection(profileName string, section *ini.Section) (*ConfigEntrySso, error) {
	configEntry := ConfigEntrySso{}
	if err := section.MapTo(&configEntry); err != nil {
		return nil, fmt.Errorf("could not parse sso config entry %s", profileName)
	}

	configEntry.Name = profileName

	return &configEntry, nil
}

func (c *Config) buildAliasProfileFromSection(profileName string, section *ini.Section) (*ConfigEntryAlias, error) {
	configEntry := ConfigEntryAlias{}

	if err := section.MapTo(&configEntry); err != nil {
		return nil, fmt.Errorf("could not parse alias config entry %s", profileName)
	}

	configEntry.Name = profileName

	return &configEntry, nil
}
