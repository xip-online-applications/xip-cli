package functions

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"os"
	"os/exec"
	"regexp"
	"xip/aws/functions/config/app"
	"xip/aws/functions/config/config"
	"xip/aws/functions/sso"
)

type Functions struct {
	AppConfiguration *app.App
	AwsConfig        *config.Config
	SsoClient        *sso.Sso
}

func New() *Functions {
	appConfig := app.New()

	if !appConfig.Initialized() {
		return nil
	}

	sess := session.Must(session.NewSession())
	awsConfig := config.New(appConfig)

	Sso := sso.New(*sess, appConfig, awsConfig)

	return &Functions{
		AppConfiguration: &appConfig,
		AwsConfig:        &awsConfig,
		SsoClient:        &Sso,
	}
}

func (f *Functions) Configure(values sso.ConfigureValues) {
	f.SsoClient.Configure(values)
}

func (f *Functions) SetDefault(profile string) {
	// Update app config
	appValues := f.AppConfiguration.Get()
	appValues.DefaultProfile = &profile
	f.AppConfiguration.Set(appValues)

	_ = os.Setenv("AWS_PROFILE", profile)
	_ = os.Setenv("AWS_DEFAULT_PROFILE", profile)
	fmt.Println("Please restart your terminal session for the profile reload to happen or run:\n\nexport AWS_DEFAULT_PROFILE=" + profile)
}

func (f *Functions) GetDefaultProfile() (*string, error) {
	prof := f.AppConfiguration.Get().DefaultProfile

	if prof == nil {
		return nil, fmt.Errorf("no default profile found")
	}

	return prof, nil
}

func (f *Functions) AddProfile(profile string, sourceProfile string, role string) {
	f.AddProfile(profile, sourceProfile, role)
}

func (f *Functions) Login(profile string) {
	f.SetDefault(profile)
	f.SsoClient.Login()
}

func (f *Functions) Identity() string {
	identity, err := exec.Command("aws", "sts", "get-caller-identity").Output()
	if err != nil {
		log.Fatal(err)
	}

	return string(identity)
}

func (f *Functions) GetAllSsoProfileNames() []string {
	config := f.SsoClient.ConfigConfig.File

	allKeys := config.Keys()
	profiles := make(map[string]string)

	re := regexp.MustCompile("^profile (\\w+)\\.sso_start_url$")

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
