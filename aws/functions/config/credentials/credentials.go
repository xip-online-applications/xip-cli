package credentials

import (
	"github.com/aws/aws-sdk-go/service/sso"
	"path/filepath"
	"time"
	"xip/utils/config_file/ini"
	"xip/utils/helpers"
)

type Credentials struct {
	File    *ini.ConfigFileIni
	Profile string
}

type Values struct {
	Region            string
	AccessKeyId       string
	SecretAccessKey   string
	SessionToken      string
	SessionExpiration time.Time
}

func New(path string, profile string) Credentials {
	config, _ := ini.New(filepath.Dir(path) + "/credentials")

	return Credentials{
		File:    config,
		Profile: profile,
	}
}

func (credentials *Credentials) Set(values Values) {
	expiration := values.SessionExpiration.Local().Format(time.RFC3339)

	_ = credentials.File.Read()
	credentials.File.Set(credentials.Profile+".region", &values.Region)
	credentials.File.Set(credentials.Profile+".aws_access_key_id", &values.AccessKeyId)
	credentials.File.Set(credentials.Profile+".aws_secret_access_key", &values.SecretAccessKey)
	credentials.File.Set(credentials.Profile+".aws_session_token", &values.SessionToken)
	credentials.File.Set(credentials.Profile+".aws_session_expiration", &expiration)
	_ = credentials.File.Write()
}

func (credentials *Credentials) FromRoleCredentials(region string, roleCredentials sso.RoleCredentials) {
	credentials.Set(Values{
		Region:            region,
		AccessKeyId:       *roleCredentials.AccessKeyId,
		SecretAccessKey:   *roleCredentials.SecretAccessKey,
		SessionToken:      *roleCredentials.SessionToken,
		SessionExpiration: helpers.IntToTime(int(*roleCredentials.Expiration / 1000)),
	})
}

func (credentials *Credentials) Get() *Values {
	_ = credentials.File.Read()

	if !credentials.File.IsSet(credentials.Profile + ".region") {
		return nil
	}

	timeParsed, _ := time.Parse(time.RFC3339, credentials.File.GetString(credentials.Profile+".aws_session_expiration"))

	return &Values{
		Region:            credentials.File.GetString(credentials.Profile + ".region"),
		AccessKeyId:       credentials.File.GetString(credentials.Profile + ".aws_access_key_id"),
		SecretAccessKey:   credentials.File.GetString(credentials.Profile + ".aws_secret_access_key"),
		SessionToken:      credentials.File.GetString(credentials.Profile + ".aws_session_token"),
		SessionExpiration: timeParsed,
	}
}

func (credentials *Credentials) Valid() bool {
	values := credentials.Get()

	if values == nil {
		return false
	}

	return len(values.SecretAccessKey) > 0 && values.SessionExpiration.After(time.Now())
}
