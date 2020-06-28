package config

import (
	"xip/aws/functions/config/app"
	"xip/utils/config_file/ini"
)

type Config struct {
	File    *ini.ConfigFileIni
	Profile *string
}

type SsoProfile struct {
	Name      string
	StartUrl  string
	Region    string
	AccountId string
	Role      string
	Output    string
}

type RoleProfile struct {
	Name          string
	SourceProfile string
	RoleArn       string
	Region        string
	Output        string
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

func (config *Config) SetSsoProfile(input SsoProfile) {
	profileName := "profile " + input.Name

	_ = config.File.Read()
	config.File.Set(profileName+".sso_start_url", &input.StartUrl)
	config.File.Set(profileName+".sso_region", &input.Region)
	config.File.Set(profileName+".sso_account_id", &input.AccountId)
	config.File.Set(profileName+".sso_role_name", &input.Role)
	config.File.Set(profileName+".region", &input.Region)
	config.File.Set(profileName+".output", &input.Output)
	_ = config.File.Write()
}

func (config *Config) GetSsoProfile(name string) *SsoProfile {
	_ = config.File.Read()

	profileName := "profile " + name

	if !config.File.IsSet(profileName + ".sso_region") {
		return nil
	}

	return &SsoProfile{
		StartUrl:  config.File.GetString(profileName + ".sso_start_url"),
		Region:    config.File.GetString(profileName + ".sso_region"),
		AccountId: config.File.GetString(profileName + ".sso_account_id"),
		Role:      config.File.GetString(profileName + ".sso_role_name"),
		Output:    config.File.GetString(profileName + ".output"),
	}
}

func (config *Config) SetRoleProfile(input RoleProfile) {
	profileName := "profile " + input.Name

	_ = config.File.Read()
	config.File.Set(profileName+".source_profile", &input.SourceProfile)
	config.File.Set(profileName+".role_arn", &input.RoleArn)
	config.File.Set(profileName+".region", &input.Region)
	config.File.Set(profileName+".output", &input.Output)
	_ = config.File.Write()
}

func (config *Config) GetRoleProfile(name string) *RoleProfile {
	_ = config.File.Read()

	profileName := "profile " + name

	if !config.File.IsSet(profileName + ".region") {
		return nil
	}

	return &RoleProfile{
		SourceProfile: config.File.GetString(profileName + ".source_profile"),
		RoleArn:       config.File.GetString(profileName + ".role_arn"),
		Region:        config.File.GetString(profileName + ".region"),
		Output:        config.File.GetString(profileName + ".output"),
	}
}

func (config *Config) Valid() bool {
	values := config.GetSsoProfile(*config.Profile)

	if values == nil {
		return false
	}

	return len(values.Region) > 0 && len(values.Role) > 0
}
