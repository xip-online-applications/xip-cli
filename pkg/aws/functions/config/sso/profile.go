package sso

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

type Profile struct {
	FileName string `json:"-"`

	AccessToken string `json:"accessToken"`
	ExpiresAt   string `json:"expiresAt"`
	Region      string `json:"region"`
	StartUrl    string `json:"startUrl"`
}

func NewProfile(AccessToken string, ExpiresAt time.Time, Region string, StartUrl string) Profile {
	usr, _ := user.Current()

	profile, err := LoadProfile(StartUrl)
	if err != nil {
		h := sha1.New()
		h.Write([]byte(StartUrl))
		sha1Hash := hex.EncodeToString(h.Sum(nil))

		profile = Profile{
			FileName: filepath.FromSlash(usr.HomeDir + "/.aws/sso/cache/" + sha1Hash + ".json"),
			StartUrl: StartUrl,
		}
	}

	profile.AccessToken = AccessToken
	profile.ExpiresAt = ExpiresAt.Local().Format(time.RFC3339)
	profile.Region = Region

	return profile
}

func LoadProfile(StartUrl string) (Profile, error) {
	usr, _ := user.Current()

	path := filepath.FromSlash(usr.HomeDir + "/.aws/sso/cache")
	files, _ := ioutil.ReadDir(path)
	for _, fileinfo := range files {
		if fileinfo.IsDir() {
			continue
		}

		profileFilePath := filepath.FromSlash(path + "/" + fileinfo.Name())
		fileContent, err := ioutil.ReadFile(profileFilePath)
		if err != nil {
			continue
		}

		var profile Profile
		if err := json.Unmarshal(fileContent, &profile); err != nil {
			continue
		}

		if len(profile.StartUrl) > 0 && profile.StartUrl == StartUrl {
			profile.FileName = profileFilePath
			return profile, nil
		}
	}

	return Profile{}, fmt.Errorf("no sso profile found")
}

func (p *Profile) Save() {
	clientJsonEncoded, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	_ = os.MkdirAll(filepath.Dir(p.FileName), 0777)
	_ = ioutil.WriteFile(p.FileName, clientJsonEncoded, 0644)
}

func (p *Profile) Delete() {
	_ = os.Remove(p.FileName)
}
