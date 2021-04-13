package config

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/service/sso"
	"gopkg.in/ini.v1"

	"xip/utils/helpers"
)

type Credentials struct {
	FileName string

	Entries map[string]CredentialsEntry
}

func NewCredentials() Credentials {
	usr, _ := user.Current()

	return Credentials{
		FileName: filepath.FromSlash(usr.HomeDir + "/.aws/credentials"),

		Entries: map[string]CredentialsEntry{},
	}
}

func LoadCredentials() (Credentials, error) {
	credentialsList := NewCredentials()
	if err := credentialsList.Load(); err != nil {
		return Credentials{}, fmt.Errorf("culd not load the credentials file: %s", err.Error())
	}

	return credentialsList, nil
}

func (c *Credentials) Load() error {
	file, err := ini.Load(c.FileName)
	if err != nil {
		file = ini.Empty()
	}

	for _, sectionName := range file.SectionStrings() {
		if sectionName == "DEFAULT" {
			continue
		}

		credentialEntry := CredentialsEntry{}
		err := file.Section(sectionName).MapTo(&credentialEntry)
		if err != nil {
			return fmt.Errorf("could not parse credentials entry %s", sectionName)
		}

		credentialEntry.Name = sectionName
		c.Entries[sectionName] = credentialEntry
	}

	return nil
}

func (c *Credentials) Save() error {
	file, err := ini.Load(c.FileName)
	if err != nil {
		file = ini.Empty()
	}

	for _, credentialEntry := range c.Entries {
		section, err := file.GetSection(credentialEntry.Name)
		if err != nil {
			section, _ = file.NewSection(credentialEntry.Name)
		}

		_ = section.ReflectFrom(&credentialEntry)
	}

	if _, err := os.Stat(c.FileName); err != nil {
		_ = os.MkdirAll(path.Dir(c.FileName), 0777)
	}

	err = file.SaveTo(c.FileName)
	if err != nil {
		return fmt.Errorf("could not save the credentials ini file to %s: %s", c.FileName, err.Error())
	}

	return nil
}

func (c *Credentials) SetFromRoleCredentials(Profile string, Region string, Credentials sso.RoleCredentials) {
	entry := CredentialsEntry{
		Name:                 Profile,
		Region:               Region,
		AwsAccessKeyId:       *Credentials.AccessKeyId,
		AwsSecretAccessKey:   *Credentials.SecretAccessKey,
		AwsSessionToken:      *Credentials.SessionToken,
		AwsSessionExpiration: helpers.IntToTime(int(*Credentials.Expiration / 1000)).Format(time.RFC3339),
	}

	c.Entries[entry.Name] = entry
	_ = c.Save()
}

func (c *Credentials) Get(Profile string) (CredentialsEntry, error) {
	if val, ok := c.Entries[Profile]; ok {
		return val, nil
	}

	return CredentialsEntry{}, nil
}

func (c *Credentials) SetDefault(Profile string) error {
	credential, err := c.Get(Profile)
	if err != nil {
		return fmt.Errorf("no credentials found for profile")
	}

	defaultCredential := credential
	defaultCredential.SetDefault()

	c.Entries[defaultCredential.Name] = defaultCredential

	return nil
}

func (c *Credentials) UnsetDefault() {
	delete(c.Entries, "default")
}
