package functions

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"os"
	"regexp"
	"xip/aws/functions/config/app"
	"xip/aws/functions/config/config"
	"xip/aws/functions/kubectl"
	"xip/aws/functions/sso"
)

type Functions struct {
	AwsSession *session.Session

	AppConfiguration *app.App
	AwsConfig        *config.Config
	SsoClient        *sso.Sso
	KubectlClient    *kubectl.Kubectl
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
	Kubectl := kubectl.New(sess, &awsConfig)

	return &Functions{
		AwsSession:       sess,
		AppConfiguration: &appConfig,
		AwsConfig:        &awsConfig,
		SsoClient:        &Sso,
		KubectlClient:    &Kubectl,
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

func (f *Functions) GetDefaultProfile() (string, error) {
	prof := f.AppConfiguration.Get().DefaultProfile

	if prof == nil {
		return "", fmt.Errorf("no default profile found")
	}

	return *prof, nil
}

func (f *Functions) AddProfile(profile string, sourceProfile string, role string) {
	source, _ := f.AwsConfig.GetSsoProfile(sourceProfile)

	f.AwsConfig.SetAliasProfile(config.AliasProfile{
		Common: config.Profile{
			Name:   profile,
			Region: source.Common.Region,
			Output: source.Common.Output,
		},
		SourceProfile: sourceProfile,
		RoleArn:       role,
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
	confFile := f.SsoClient.ConfigConfig.File

	allKeys := confFile.Keys()
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

func (f *Functions) RegisterKubectlProfile(clusterName string, roleArn string, profile string, namespace string, alias string) error {
	return f.KubectlClient.RegisterProfile(clusterName, roleArn, profile, namespace, alias)
}
