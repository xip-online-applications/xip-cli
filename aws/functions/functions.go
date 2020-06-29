package functions

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

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

	awsConfig, _ := config.LoadConfig()
	Sso := sso.New(*sess, appConfig)
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
	awsConfig, _ := config.LoadConfig()
	source, _ := awsConfig.GetProfile(sourceProfile)

	err := awsConfig.SetAliasProfile(profile, source.Region, source.Output, sourceProfile, role)
	if err != nil {
		panic(err)
	}

	f.Login("")
	f.SetDefault(profile)
}

func (f *Functions) Login(profile string) {
	if len(profile) > 1 {
		f.SsoClient.Login(profile)
		f.SetDefault(profile)
	} else {
		for _, value := range f.GetAllProfileNames() {
			f.SsoClient.Login(value)
		}
	}
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

func (f *Functions) GetAllProfileNames() []string {
	configFile, _ := config.LoadConfig()
	var profiles []string

	for _, value := range configFile.SsoEntries {
		profiles = append(profiles, value.Name)
	}

	for _, value := range configFile.AliasEntries {
		profiles = append(profiles, value.Name)
	}

	return profiles
}

func (f *Functions) RegisterKubectlProfile(clusterName string, roleArn string, profile string, namespace string, alias string) error {
	return f.KubectlClient.RegisterProfile(clusterName, roleArn, profile, namespace, alias)
}
