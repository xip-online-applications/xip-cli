package cli

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type SsoProfile struct {
	FileName string `json:"-"`

	ProviderType string      `json:"ProviderType"`
	Credentials  Credentials `json:"Credentials"`
}

func NewSsoProfile(FileName string, AccessKeyId string, SecretAccessKey string, SessionToken string, Expiration time.Time) SsoProfile {
	usr, _ := user.Current()

	return SsoProfile{
		FileName: usr.HomeDir + "/.aws/cli/cache/" + FileName + ".json",

		ProviderType: "sso",
		Credentials: Credentials{
			AccessKeyId:     AccessKeyId,
			SecretAccessKey: SecretAccessKey,
			SessionToken:    SessionToken,
			Expiration:      Expiration.Local().Format(time.RFC3339),
		},
	}
}

func CreateSsoProfileFileName(AccountId string, RoleName string, StartUrl string) string {
	name, _ := json.MarshalIndent(struct {
		AccountId string `json:"accountId"`
		RoleName  string `json:"roleName"`
		StartUrl  string `json:"startUrl"`
	}{
		AccountId: AccountId,
		RoleName:  RoleName,
		StartUrl:  StartUrl,
	}, "", "")

	nameString := strings.ReplaceAll(string(name), "\n", "")

	h := sha1.New()
	h.Write([]byte(nameString))
	return hex.EncodeToString(h.Sum(nil))
}

func (sp *SsoProfile) Save() {
	clientJsonEncoded, err := json.Marshal(sp)
	if err != nil {
		panic(err)
	}

	_ = os.MkdirAll(filepath.Dir(sp.FileName), 0777)
	_ = ioutil.WriteFile(sp.FileName, clientJsonEncoded, 0644)
}
