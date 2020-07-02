package functions

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"xip/aws/functions/config/app"
	"xip/aws/functions/config/config"
	"xip/aws/functions/eks"
	"xip/aws/functions/kubectl"
	"xip/aws/functions/sso"
)

type Functions struct {
	AwsSession *session.Session

	AppConfiguration *app.App
	AwsConfig        *config.Config
	SsoClient        *sso.Sso
	EksClient        *eks.Eks
	KubectlClient    *kubectl.Kubectl
}

func New() Functions {
	f := Functions{}
	f.setup()

	return f
}

func (f *Functions) Configure(values sso.ConfigureValues) {
	f.SsoClient.Configure(values)
}

func (f *Functions) SetDefault(profile string) {
	// Update app config
	appValues := f.AppConfiguration.Get()
	appValues.DefaultProfile = &profile
	f.AppConfiguration.Set(appValues)

	// Update AWS session information
	_ = os.Setenv("AWS_PROFILE", profile)
	_ = os.Setenv("AWS_DEFAULT_PROFILE", profile)

	// Reload the profile stuff
	f.setup()
}

func (f *Functions) PrintDefaultHelp(Profile string) {
	fmt.Println("Please restart your terminal session for the profile reload to happen or run:")
	fmt.Println("")
	fmt.Println("export AWS_PROFILE=" + Profile)
	fmt.Println("export AWS_DEFAULT_PROFILE=" + Profile)
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

	f.SetDefault(sourceProfile)
	f.Login("")
	f.SetDefault(profile)
}

func (f *Functions) Login(profile string) {
	if len(profile) > 1 {
		f.SsoClient.Login(profile)
		f.SetDefault(profile)
	} else {
		currentDefault, _ := f.GetDefaultProfile()
		defer f.SetDefault(currentDefault)

		for _, value := range f.GetAllProfileNames() {
			f.SetDefault(value)
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

func (f *Functions) GetEksToken(profile string, clusterName string, roleArn string) (string, string, error) {
	prof, err := f.AwsConfig.GetProfile(profile)
	if err != nil {
		return "", "", err
	}

	currentDefaultProfile, _ := f.GetDefaultProfile()
	defer f.SetDefault(currentDefaultProfile)

	f.SetDefault(profile)
	return f.EksClient.GetToken(prof.Region, clusterName, roleArn)
}

func (f *Functions) setup() {
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
	EksClient := eks.New(*sess)
	Kubectl := kubectl.New(sess, &awsConfig)

	f.AwsSession = sess
	f.AppConfiguration = &appConfig
	f.AwsConfig = &awsConfig
	f.SsoClient = &Sso
	f.EksClient = &EksClient
	f.KubectlClient = &Kubectl
}
