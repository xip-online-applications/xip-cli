package config

import (
	"fmt"
	"xip/aws/functions/config/app"
	"xip/utils/config_file/ini"
)

type Config struct {
	File    *ini.ConfigFileIni
	Profile *string
}

type Profile struct {
	Name   string
	Region string
	Output string
}

type SsoProfile struct {
	Common Profile

	StartUrl  string
	AccountId string
	Role      string
	SsoRegion string
}

type AliasProfile struct {
	Common Profile

	SourceProfile string
	RoleArn       string
}

func New(appConfig app.App) Config {
	var config *ini.ConfigFileIni
	path := appConfig.Get().AwsConfigPath

	if path != nil {
		config, _ = ini.New(*path)
	}

	return Config{
		File:    config,
		Profile: appConfig.Get().DefaultProfile,
	}
}

func (config *Config) GetProfile(name string) *Profile {
	_ = config.File.Read()

	profileName := "profile " + name
	if !config.File.IsSet(profileName + ".region") {
		return nil
	}

	return &Profile{
		Name:   name,
		Region: config.File.GetString(profileName + ".region"),
		Output: config.File.GetString(profileName + ".output"),
	}
}

func (config *Config) SetSsoProfile(input SsoProfile) {
	profileName := "profile " + input.Common.Name

	_ = config.File.Read()
	config.File.Set(profileName+".sso_start_url", &input.StartUrl)
	config.File.Set(profileName+".sso_region", &input.SsoRegion)
	config.File.Set(profileName+".sso_account_id", &input.AccountId)
	config.File.Set(profileName+".sso_role_name", &input.Role)
	config.File.Set(profileName+".region", &input.Common.Region)
	config.File.Set(profileName+".output", &input.Common.Output)
	_ = config.File.Write()
}

func (config *Config) GetSsoProfile(name string) (SsoProfile, error) {
	common := config.GetProfile(name)
	if common == nil {
		return SsoProfile{}, fmt.Errorf("profile " + name + " not found")
	}

	profileName := "profile " + name
	if !config.File.IsSet(profileName + ".sso_region") {
		return SsoProfile{}, fmt.Errorf("profile " + name + " not found")
	}

	return SsoProfile{
		Common:    *common,
		StartUrl:  config.File.GetString(profileName + ".sso_start_url"),
		SsoRegion: config.File.GetString(profileName + ".sso_region"),
		AccountId: config.File.GetString(profileName + ".sso_account_id"),
		Role:      config.File.GetString(profileName + ".sso_role_name"),
	}, nil
}

func (config *Config) SetAliasProfile(input AliasProfile) {
	profileName := "profile " + input.Common.Name

	_ = config.File.Read()
	config.File.Set(profileName+".source_profile", &input.SourceProfile)
	config.File.Set(profileName+".role_arn", &input.RoleArn)
	config.File.Set(profileName+".region", &input.Common.Region)
	config.File.Set(profileName+".output", &input.Common.Output)
	_ = config.File.Write()
}

func (config *Config) GetAliasProfile(name string) *AliasProfile {
	common := config.GetProfile(name)
	if common == nil {
		return nil
	}

	profileName := "profile " + name
	if !config.File.IsSet(profileName + ".source_profile") {
		return nil
	}

	return &AliasProfile{
		Common:        *common,
		SourceProfile: config.File.GetString(profileName + ".source_profile"),
		RoleArn:       config.File.GetString(profileName + ".role_arn"),
	}
}

func (config *Config) Valid() bool {
	values, err := config.GetSsoProfile(*config.Profile)
	if err != nil {
		return false
	}

	return len(values.Common.Region) > 0 && len(values.Role) > 0
}
