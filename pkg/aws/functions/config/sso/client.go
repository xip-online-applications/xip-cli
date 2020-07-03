package sso

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"xip/utils/helpers"
)

type Client struct {
	FileName string `json:"-"`

	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	ExpiresAt    string `json:"expiresAt"`
}

func NewClient(ClientId string, ClientSecret string, ClientExpiration time.Time) Client {
	usr, _ := user.Current()

	client, err := LoadClient()
	if err != nil {
		client = Client{
			FileName: usr.HomeDir + "/.aws/sso/cache/botocore-sso-id-eu-west-1.json",
		}
	}

	client.ClientId = ClientId
	client.ClientSecret = ClientSecret
	client.ExpiresAt = ClientExpiration.Local().Format(time.RFC3339)

	return client
}

func LoadClient() (Client, error) {
	usr, _ := user.Current()

	fileContent, err := ioutil.ReadFile(usr.HomeDir + "/.aws/sso/cache/botocore-sso-id-eu-west-1.json")
	if err != nil {
		return Client{}, fmt.Errorf("could not read client file")
	}

	var client Client
	if err := json.Unmarshal(fileContent, &client); err != nil {
		return Client{}, fmt.Errorf("could not decode client: %s", err.Error())
	}

	return client, nil
}

func (c *Client) Save() {
	clientJsonEncoded, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	_ = os.MkdirAll(filepath.Dir(c.FileName), 0777)
	_ = ioutil.WriteFile(c.FileName, clientJsonEncoded, 0644)
}

func (c *Client) Valid() bool {
	return len(c.ExpiresAt) > 0 && helpers.StringToTime(c.ExpiresAt).After(time.Now())
}
