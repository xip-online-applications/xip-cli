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

	"github.com/aws/aws-sdk-go/service/sts"
)

type RoleProfile struct {
	FileName string `json:"-"`

	AssumedRoleUser AssumedRole `json:"AssumedRoleUser"`
	Credentials     Credentials `json:"Credentials"`
}

func NewRoleProfile(FileName string, AssumeRoleOutput sts.AssumeRoleOutput) RoleProfile {
	usr, _ := user.Current()

	return RoleProfile{
		FileName: usr.HomeDir + "/.aws/cli/cache/" + FileName + ".json",

		AssumedRoleUser: AssumedRole{
			Arn:           *AssumeRoleOutput.AssumedRoleUser.Arn,
			AssumedRoleId: *AssumeRoleOutput.AssumedRoleUser.AssumedRoleId,
		},
		Credentials: Credentials{
			AccessKeyId:     *AssumeRoleOutput.Credentials.AccessKeyId,
			SecretAccessKey: *AssumeRoleOutput.Credentials.SecretAccessKey,
			SessionToken:    *AssumeRoleOutput.Credentials.SessionToken,
			Expiration:      AssumeRoleOutput.Credentials.Expiration.Local().Format(time.RFC3339),
		},
	}
}

func CreateRoleProfileFileName(RoleArn string, DurationSeconds int64) string {
	name, _ := json.MarshalIndent(struct {
		RoleArn string `json:"RoleArn"`
	}{
		RoleArn: RoleArn,
	}, "", "")

	nameString := strings.ReplaceAll(string(name), "\n", "")

	h := sha1.New()
	h.Write([]byte(nameString))
	return hex.EncodeToString(h.Sum(nil))
}

func (rp *RoleProfile) Save() {
	clientJsonEncoded, err := json.Marshal(rp)
	if err != nil {
		panic(err)
	}

	_ = os.MkdirAll(filepath.Dir(rp.FileName), 0777)
	_ = ioutil.WriteFile(rp.FileName, clientJsonEncoded, 0644)
}
