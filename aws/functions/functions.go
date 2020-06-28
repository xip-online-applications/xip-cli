package functions

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"os"
	"regexp"
	"xip/aws/functions/config/app"
	"xip/aws/functions/config/config"
	"xip/aws/functions/sso"
)

type Functions struct {
	AwsSession *session.Session

	AppConfiguration *app.App
	AwsConfig        *config.Config
	SsoClient        *sso.Sso
}

func New() *Functions {
	appConfig := app.New()

	if !appConfig.Initialized() {
		return nil
	}

	var (
		prof    = appConfig.Get().DefaultProfile
		profile = ""
	)

	if prof != nil {
		profile = *prof
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))

	awsConfig := config.New(appConfig)

	Sso := sso.New(*sess, appConfig, awsConfig)

	return &Functions{
		AwsSession:       sess,
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
	source := f.AwsConfig.GetSsoProfile(sourceProfile)

	f.AwsConfig.SetRoleProfile(config.RoleProfile{
		Name:          profile,
		SourceProfile: sourceProfile,
		RoleArn:       role,
		Region:        source.Region,
		Output:        source.Output,
	})
}

func (f *Functions) Login(profile string) {
	creds := f.SsoClient.CredentialsConfig.ForProfile(profile)
	if creds == nil {
		panic("no credentials found for given profile")
	}

	f.SetDefault(profile)
	f.SsoClient.Login(*creds)
}

func (f *Functions) Identity() string {
	stsClient := sts.New(f.AwsSession)

	input := &sts.GetCallerIdentityInput{}
	identity, err := stsClient.GetCallerIdentity(input)

	if err != nil {
		panic("could not retrieve identity")
	}

	return *identity.UserId
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
