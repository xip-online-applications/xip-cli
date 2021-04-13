package config

type CredentialsEntry struct {
	Name   string `ini:"-"`
	Parent string `ini:"parent,omitempty"`

	Region               string `ini:"region"`
	AwsAccessKeyId       string `ini:"aws_access_key_id"`
	AwsSecretAccessKey   string `ini:"aws_secret_access_key"`
	AwsSessionToken      string `ini:"aws_session_token,omitempty"`
	AwsSessionExpiration string `ini:"aws_session_expiration,omitempty"`
}

func (cee *CredentialsEntry) SetDefault() {
	cee.Parent = cee.Name
	cee.Name = "default"
}
