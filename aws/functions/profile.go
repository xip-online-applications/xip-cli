package functions

import (
	"fmt"
	"log"
	"os/exec"
	"xip/utils/ini_config_file"
)

func SetDefault(path string, profile string) {
	config := _GetConfig(path)

	if !config.IsSet("profile " + profile + ".output") {
		panic(fmt.Errorf("Profile %s probably does not exist\n", profile))
	}

	appConf := ini_config_file.AppConf()
	appConf.Set("aws.default_profile", profile)
	_ = appConf.Write()

	fmt.Println("Please restart your terminal session for the profile reload to happen or run:\n\nexport AWS_DEFAULT_PROFILE=" + profile)
}

func GetDefault() string {
	appConf := ini_config_file.AppConf()
	return appConf.GetString("aws.default_profile")
}

func CreateOrUpdateSsoProfile(path string, profile string, role string, region string) {
	config := _GetConfig(path)

	profileName := "profile " + profile

	config.Set(profileName+".sso_start_url", "https://xip.awsapps.com/start")
	config.Set(profileName+".sso_region", "eu-west-1")
	config.Set(profileName+".sso_account_id", "616582671099")
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

func _GetConfig(path string) *ini_config_file.ConfigFileIni {
	config, err := ini_config_file.New(path)
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	return config
}
