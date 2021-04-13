package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

type EntryConfig struct {
	Name   string `ini:"-"`
	Parent string `ini:"parent,omitempty"`

	Region string `ini:"region"`
	Output string `ini:"output"`

	StartUrl  string `ini:"sso_start_url,omitempty"`
	AccountId string `ini:"sso_account_id,omitempty"`
	Role      string `ini:"sso_role_name,omitempty"`
	SsoRegion string `ini:"sso_region,omitempty"`

	SourceProfile   string `ini:"source_profile,omitempty"`
	RoleArn         string `ini:"role_arn,omitempty"`
	ExternalId      string `ini:"external_id,omitempty"`
	RoleSessionName string `ini:"role_session_name,omitempty"`
}

func EntryFromSection(profileName string, section *ini.Section) (*EntryConfig, error) {
	configEntry := EntryConfig{}

	if err := section.MapTo(&configEntry); err != nil {
		return nil, fmt.Errorf("could not parse alias config entry %s", profileName)
	}

	configEntry.Name = profileName

	return &configEntry, nil
}

func (ec *EntryConfig) IsSsoProfile() bool {
	return ec.StartUrl != "" && ec.AccountId != "" && ec.Role != "" && ec.SsoRegion != ""
}

func (ec *EntryConfig) IsAliasProfile() bool {
	return ec.SourceProfile != "" && ec.RoleArn != ""
}

func (ec *EntryConfig) IsRegularProfile() bool {
	return !ec.IsSsoProfile() && !ec.IsAliasProfile()
}

func (ec *EntryConfig) SetIsDefaultProfile() {
	ec.Parent = ec.Name
	ec.Name = "default"

	if ec.IsAliasProfile() {
		ec.RoleSessionName = "xip-cli--" + ec.Name
	}
}

func (ec *EntryConfig) IsDefaultProfile() bool {
	return ec.Name == "default"
}
