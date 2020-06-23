package functions

import (
	json2 "encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
	"xip/utils/config_file/ini"
	"xip/utils/config_file/json"
)

func SetDefault(path string, profile string) {
	config := _GetConfig(path)

	if !config.IsSet("profile " + profile + ".output") {
		panic(fmt.Errorf("Profile %s probably does not exist\n", profile))
	}

	appConf := ini.AppConf()
	appConf.Set("aws.default_profile", profile)
	_ = appConf.Write()

	fmt.Println("Please restart your terminal session for the profile reload to happen or run:\n\nexport AWS_DEFAULT_PROFILE=" + profile)
}

func GetDefaultProfile() string {
	appConf := ini.AppConf()
	return appConf.GetString("aws.default_profile")
}

func CreateOrUpdateSsoProfile(path string, profile string, role string, region string, startUrl string, accountId string) {
	config := _GetConfig(path)

	profileName := "profile " + profile

	config.Set(profileName+".sso_start_url", startUrl)
	config.Set(profileName+".sso_region", region)
	config.Set(profileName+".sso_account_id", accountId)
	config.Set(profileName+".sso_role_name", role)
	config.Set(profileName+".region", region)
	config.Set(profileName+".output", "json")

	if err := config.Write(); err != nil {
		panic(fmt.Errorf("Fatal error writing config file: %s \n", err))
	}
}

func CreateOrUpdateRoleAssumeProfile(path string, profile string, sourceProfile string, role string) {
	config := _GetConfig(path)

	if !config.IsSet("profile " + sourceProfile + ".output") {
		panic(fmt.Errorf("Source profile %s probably does not exist\n", sourceProfile))
	}

	profileName := "profile " + profile

	config.Set(profileName+".role_arn", role)
	config.Set(profileName+".source_profile", sourceProfile)
	config.Set(profileName+".region", config.GetString("profile "+sourceProfile+".region"))
	config.Set(profileName+".output", config.GetString("profile "+sourceProfile+".output"))

	if err := config.Write(); err != nil {
		panic(fmt.Errorf("Fatal error writing config file: %s \n", err))
	}
}

func Login(profile string) {
	_, err := exec.Command("aws", "sso", "login", "--profile", profile).Output()
	if err != nil {
		log.Fatal(err)
	}
}

func Sync(path string, profile string) {
	config := _GetConfig(path)

	var (
		accessKeyId     string
		secretAccessKey string
		sessionToken    string
		expirationTime  string
	)

	if config.IsSet("profile " + profile + ".sso_region") {
		region := config.GetString("profile " + profile + ".sso_region")
		startUrl := config.GetString("profile " + profile + ".sso_start_url")
		accessToken := _GetSsoAccessToken(path, startUrl, region)

		ssoCreds, err := exec.Command(
			"aws", "sso", "get-role-credentials",
			"--output", "json",
			"--profile", profile,
			"--region", region,
			"--role-name", config.GetString("profile "+profile+".sso_role_name"),
			"--account-id", config.GetString("profile "+profile+".sso_account_id"),
			"--access-token", accessToken,
		).Output()
		if err != nil {
			log.Fatal(err)
		}

		var ssoCredsMapped map[string]map[string]interface{}
		if err := json2.Unmarshal(ssoCreds, &ssoCredsMapped); err != nil {
			return
		}

		accessKeyId = fmt.Sprintf("%v", ssoCredsMapped["roleCredentials"]["accessKeyId"])
		secretAccessKey = fmt.Sprintf("%v", ssoCredsMapped["roleCredentials"]["secretAccessKey"])
		sessionToken = fmt.Sprintf("%v", ssoCredsMapped["roleCredentials"]["sessionToken"])

		t1, _ := strconv.ParseFloat(fmt.Sprintf("%f", ssoCredsMapped["roleCredentials"]["expiration"]), 32)
		t2 := int64(t1 / 1000)
		t3 := time.Unix(t2, 0)

		expirationTime = t3.Local().Format(time.RFC3339)
	} else {
		roleArn := config.GetString("profile " + profile + ".role_arn")

		ssoCreds, err := exec.Command(
			"aws", "sts", "assume-role",
			"--output", "json",
			"--profile", profile,
			"--role-arn", roleArn,
			"--role-session-name", "tm",
			"--duration-seconds", "3600",
		).Output()
		if err != nil {
			log.Fatal(err)
		}

		var ssoCredsMapped map[string]map[string]interface{}
		if err := json2.Unmarshal(ssoCreds, &ssoCredsMapped); err != nil {
			return
		}

		accessKeyId = fmt.Sprintf("%v", ssoCredsMapped["Credentials"]["AccessKeyId"])
		secretAccessKey = fmt.Sprintf("%v", ssoCredsMapped["Credentials"]["SecretAccessKey"])
		sessionToken = fmt.Sprintf("%v", ssoCredsMapped["Credentials"]["SessionToken"])
		expirationTime = fmt.Sprintf("%v", ssoCredsMapped["Credentials"]["Expiration"])

		t1, _ := time.Parse(time.RFC3339, expirationTime)
		expirationTime = t1.Local().Format(time.RFC3339)
	}

	credentials := _GetConfig(filepath.Dir(path) + "/credentials")
	credentials.Set(profile+".region", config.GetString("profile "+profile+".region"))
	credentials.Set(profile+".aws_access_key_id", accessKeyId)
	credentials.Set(profile+".aws_secret_access_key", secretAccessKey)
	credentials.Set(profile+".aws_session_token", sessionToken)
	credentials.Set(profile+".aws_session_expiration", expirationTime)

	if err := credentials.Write(); err != nil {
		panic(err)
	}

	fmt.Println("Please restart your terminal session for the profile reload to happen or run:\n\nexport AWS_PROFILE=" + profile)
}

func Identity() string {
	identity, err := exec.Command("aws", "sts", "get-caller-identity").Output()
	if err != nil {
		log.Fatal(err)
	}

	return string(identity)
}

func GetAllProfileNames(path string) []string {
	config := _GetConfig(path)

	allKeys := config.Keys()
	profiles := make(map[string]string)

	re := regexp.MustCompile("^profile (\\w+)\\..+$")

	for _, value := range allKeys {
		val := re.FindStringSubmatch(value)

		if len(val) < 2 {
			continue
		}

		if _, ok := profiles[val[1]]; ok {
			continue
		}

		profiles[val[1]] = val[1]
	}

	var keys []string
	for k := range profiles {
		keys = append(keys, k)
	}

	return keys
}

func _GetConfig(path string) *ini.ConfigFileIni {
	config, err := ini.New(path)
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	return config
}

func _GetJsonConfig(path string) *json.ConfigFileJson {
	config, err := json.New(path)
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	return config
}

func _GetSsoAccessToken(root string, startUrl string, region string) string {
	var accessToken string

	_ = filepath.Walk(filepath.Dir(root)+"/sso/cache", func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".json" {
			return nil
		}

		file, _ := ioutil.ReadFile(path)

		var raw map[string]interface{}
		if err := json2.Unmarshal(file, &raw); err != nil {
			return nil
		}

		if _, ok := raw["accessToken"]; !ok {
			return nil
		}

		if _, ok := raw["startUrl"]; !ok {
			return nil
		}

		if _, ok := raw["region"]; !ok {
			return nil
		}

		if raw["region"] == region && raw["startUrl"] == startUrl {
			accessToken = fmt.Sprintf("%v", raw["accessToken"])
		}

		return nil
	})

	return accessToken
}
