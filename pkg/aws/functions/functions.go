package functions

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"

	"xip/aws/functions/config/config"
	"xip/aws/functions/eks"
	"xip/aws/functions/kubectl"
	"xip/aws/functions/sso"
)

type Functions struct {
	AwsSession *session.Session

	AwsConfig     config.Config
	SsoClient     *sso.Sso
	EksClient     *eks.Eks
	EcrClient     *ecr.ECR
	KubectlClient *kubectl.Kubectl
}

func New() Functions {
	f := Functions{}
	f.setup()

	return f
}

func (f *Functions) Configure(values sso.ConfigureValues) {
	f.SsoClient.Configure(values)
}

func (f *Functions) SetDefault(profile string) error {
	credentials, err := config.LoadCredentials()
	if err != nil {
		return err
	}

	prof, _ := f.AwsConfig.SetDefaultProfile(profile)
	_ = f.AwsConfig.Save()

	if prof.IsRegularProfile() {
		_ = credentials.SetDefault(profile)
	} else {
		credentials.UnsetDefault()
	}
	_ = credentials.Save()

	// Update AWS session information
	_ = os.Setenv("AWS_PROFILE", profile)
	_ = os.Setenv("AWS_DEFAULT_PROFILE", profile)

	// Reload the profile stuff
	f.setup()

	return nil
}

func (f *Functions) PrintDefaultHelp(Profile string) {
	fmt.Println("Please restart your terminal session for the profile reload to happen or run:")
	fmt.Println("")
	fmt.Println("export AWS_PROFILE=" + Profile)
	fmt.Println("export AWS_DEFAULT_PROFILE=" + Profile)
}

func (f *Functions) GetDefaultProfile() (string, error) {
	prof := f.AwsConfig.DefaultProfile

	if prof == nil {
		return "", fmt.Errorf("no default profile found")
	}

	return prof.Name, nil
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
		f.SetDefault(profile)
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

	for _, value := range configFile.ConfigEntries {
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

	f.Login(profile)
	return f.EksClient.GetToken(prof.Region, clusterName, roleArn)
}

func (f *Functions) GetEcrPassword(profile string) (string, error) {
	currentDefaultProfile, _ := f.GetDefaultProfile()
	defer f.SetDefault(currentDefaultProfile)

	f.Login(profile)
	input := &ecr.GetAuthorizationTokenInput{}

	result, err := f.EcrClient.GetAuthorizationToken(input)
	if err != nil {
		return "", err
	}

	authData := result.AuthorizationData
	if len(authData) == 0 {
		return "", nil
	}

	authToken := authData[0].AuthorizationToken
	decodedAuthToken, _ := base64.StdEncoding.DecodeString(*authToken)

	decodedAuthTokenString := string(decodedAuthToken)
	splittedAuthString := strings.Split(decodedAuthTokenString, ":")

	return splittedAuthString[1], nil
}

func (f *Functions) setup() {
	f.AwsConfig, _ = config.LoadConfig()

	var (
		prof    = f.AwsConfig.DefaultProfile
		profile = ""
	)

	if prof != nil {
		profile = prof.Name
	}

	f.AwsSession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))

	Sso := sso.New(*f.AwsSession)
	EksClient := eks.New(*f.AwsSession)
	Kubectl := kubectl.New(f.AwsSession, f.AwsConfig)

	f.SsoClient = &Sso
	f.EksClient = &EksClient
	f.EcrClient = ecr.New(f.AwsSession)
	f.KubectlClient = &Kubectl
}
