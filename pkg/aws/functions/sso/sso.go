package sso

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	awssso "github.com/aws/aws-sdk-go/service/sso"
	"github.com/aws/aws-sdk-go/service/ssooidc"
	"github.com/aws/aws-sdk-go/service/sts"

	"xip/aws/functions/config/cli"
	"xip/aws/functions/config/config"
	"xip/aws/functions/config/sso"
	"xip/utils/helpers"
)

type ConfigureValues struct {
	Region    *string
	StartUrl  *string
	Profile   *string
	AccountId *string
	RoleName  *string
}

type Sso struct {
	// AWS information
	awsSession    *session.Session
	ssoClient     *awssso.SSO
	ssoOidcClient *ssooidc.SSOOIDC
	stsClient     *sts.STS

	// On-the-fly information
	deviceCodeExpiration *int32
	retryCount           int8
}

type LoginOptions struct {
	Session *session.Session
	Sso     *awssso.SSO
	SsoOidc *ssooidc.SSOOIDC
	Sts     *sts.STS
	Config  *config.Config

	RetryCount int
}

// https://docs.aws.amazon.com/cognito/latest/developerguide/token-endpoint.html
// https://docs.aws.amazon.com/singlesignon/latest/OIDCAPIReference/API_CreateToken.html
// https://docs.aws.amazon.com/singlesignon/latest/PortalAPIReference/API_GetRoleCredentials.html

func New(awsSession session.Session) Sso {
	s := Sso{
		awsSession: &awsSession,
	}

	s.load()

	return s
}

func (s *Sso) Login(Profile string, AllOptions ...*LoginOptions) {
	options := &LoginOptions{}

	// Config
	awsConfig, _ := config.LoadConfig()
	profile, _ := awsConfig.GetProfile(Profile)

	if len(AllOptions) > 1 {
		panic("Only one options entry possible")
	} else if len(AllOptions) == 1 {
		options = AllOptions[0]
	} else {
		// Setup login options
		options = &LoginOptions{}
		options.Config = &awsConfig
		options.Session = s.awsSession.Copy(aws.NewConfig().WithRegion(profile.Region).WithSTSRegionalEndpoint(endpoints.RegionalSTSEndpoint))
		options.Sso = awssso.New(options.Session, options.Session.Config)
		options.SsoOidc = ssooidc.New(options.Session, options.Session.Config)
		options.Sts = sts.New(options.Session, options.Session.Config)
	}

	// Retry mechanism
	defer func() {
		if err := recover(); err == nil {
			return
		}

		if options.RetryCount > 3 {
			err := recover()
			_, _ = fmt.Fprintf(os.Stderr, "Failed too many times: %s\n", err)
			os.Exit(1)
		}

		options.RetryCount++
		s.Login(Profile, options)
	}()

	// Register the device if needed
	s.registerClient(options)

	if s.isAliasProfile(options, Profile) {
		// Just assume the role
		s.assumeRole(options, Profile)
	} else {
		// Authorize the device
		s.authorizeDevice(options, Profile)

		// Retrieve the role credentials by assuming it
		s.retrieveRoleCredentials(options, Profile)
	}
}

func (s *Sso) Configure(values ConfigureValues) {
	// Get config file
	awsConfig, _ := config.LoadConfig()
	err := awsConfig.SetSsoProfile(*values.Profile, *values.Region, "json", *values.StartUrl, *values.AccountId, *values.RoleName, *values.Region)
	if err != nil {
		panic(err)
	}

	_, _ = awsConfig.SetDefaultProfile(*values.Profile)

	// Reload the configuration
	s.load()

	// Register the device if needed
	s.Login(*values.Profile)
}

func (s *Sso) load() {
	// // App configuration
	// clientValues := s.appConfig.Get()
	//
	// // Try to load default profile
	// if profile := clientValues.DefaultProfile; profile != nil {
	// 	awsConfig, _ := config.LoadConfig()
	// 	awsProfile, _ := awsConfig.GetProfile(*profile)
	// 	s.awsSession.Config.Region = &awsProfile.Region
	// }

	// Create a SSOOIDC app with additional configuration
	s.ssoClient = awssso.New(s.awsSession, s.awsSession.Config)
	s.ssoOidcClient = ssooidc.New(s.awsSession, s.awsSession.Config)
	s.stsClient = sts.New(s.awsSession, s.awsSession.Config)
}

func (s *Sso) registerClient(options *LoginOptions) {
	awsConfig, _ := sso.LoadClient()
	if awsConfig.Valid() {
		return
	}

	clientName := "xip_cli_tool"
	clientType := "public"

	clientInput := &ssooidc.RegisterClientInput{
		ClientName: &clientName,
		ClientType: &clientType,
		Scopes:     nil,
	}

	output, err := options.SsoOidc.RegisterClient(clientInput)
	if err != nil {
		panic(err)
	}

	expiration := helpers.IntToTime(int(*output.ClientSecretExpiresAt))

	awsConfig = sso.NewClient(*output.ClientId, *output.ClientSecret, expiration)
	awsConfig.Save()
}

func (s *Sso) authorizeDevice(options *LoginOptions, Profile string) {
	if s.hasValidSsoProfileWithAccessToken(Profile) {
		return
	}

	if s.deviceCodeExpiration != nil && *s.deviceCodeExpiration > int32(time.Now().Unix()) {
		return
	}

	awsConfig, _ := config.LoadConfig()
	awsProfile, err := awsConfig.GetSsoProfile(Profile)
	if err != nil {
		return
	}

	awsClientConfig, _ := sso.LoadClient()

	clientInput := &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     &awsClientConfig.ClientId,
		ClientSecret: &awsClientConfig.ClientSecret,
		StartUrl:     &awsProfile.StartUrl,
	}

	output, err := options.SsoOidc.StartDeviceAuthorization(clientInput)
	if err != nil {
		panic(err)
	}

	if err = helpers.OpenBrowser(*output.VerificationUriComplete); err != nil {
		fmt.Println("Could not open your browser, please visit this url:")
		fmt.Println(*output.VerificationUriComplete)
	}

	tokenExpiration := int32(time.Now().Unix()) + int32(*output.ExpiresIn)
	s.deviceCodeExpiration = &tokenExpiration

	retryCount := int(*output.ExpiresIn) / int(*output.Interval)
	sleepTimeout, _ := time.ParseDuration(fmt.Sprintf("%ds", *output.Interval))

	for i := 0; i < retryCount; i++ {
		if err := s.createToken(options, Profile, awsClientConfig.ClientId, awsClientConfig.ClientSecret, *output.DeviceCode, *output.UserCode); err == nil {
			break
		}

		time.Sleep(sleepTimeout)
	}
}

func (s *Sso) createToken(options *LoginOptions, Profile string, ClientId string, ClientSecret string, DeviceCode string, UserCode string) error {
	grantType := "urn:ietf:params:oauth:grant-type:device_code"

	clientInput := &ssooidc.CreateTokenInput{
		ClientId:     &ClientId,
		ClientSecret: &ClientSecret,
		DeviceCode:   &DeviceCode,
		Code:         &UserCode,
		GrantType:    &grantType,
	}

	output, err := options.SsoOidc.CreateToken(clientInput)
	if err != nil {
		return err
	}

	expiration := helpers.IntToTime(int(time.Now().Unix() + *output.ExpiresIn))

	awsConfig, _ := config.LoadConfig()
	awsSsoProfile, _ := awsConfig.GetSsoProfile(Profile)

	awsCredentials := sso.NewProfile(*output.AccessToken, expiration, awsSsoProfile.Region, awsSsoProfile.StartUrl)
	awsCredentials.Save()

	return nil
}

func (s *Sso) retrieveRoleCredentials(options *LoginOptions, Profile string) {
	awsConfig, _ := config.LoadConfig()
	awsSsoProfile, _ := awsConfig.GetSsoProfile(Profile)

	ssoProfile, _ := sso.LoadProfile(awsSsoProfile.StartUrl)

	input := awssso.GetRoleCredentialsInput{
		AccessToken: &ssoProfile.AccessToken,
		AccountId:   &awsSsoProfile.AccountId,
		RoleName:    &awsSsoProfile.Role,
	}

	output, err := options.Sso.GetRoleCredentials(&input)
	if err != nil {
		ssoProfile.Delete()
		panic(err)
	}

	awsCredentials, _ := config.LoadCredentials()
	awsCredentials.SetFromRoleCredentials(Profile, awsSsoProfile.Region, *output.RoleCredentials)
	awsCredential, _ := awsCredentials.Get(Profile)

	fileName := cli.CreateSsoProfileFileName(awsSsoProfile.AccountId, awsSsoProfile.Role, awsSsoProfile.StartUrl)
	ssoClientProfile := cli.NewSsoProfile(fileName, awsCredential.AwsAccessKeyId, awsCredential.AwsSecretAccessKey, awsCredential.AwsSessionToken, helpers.StringToTime(awsCredential.AwsSessionExpiration))
	ssoClientProfile.Save()
}

func (s *Sso) assumeRole(options *LoginOptions, Profile string) {
	awsConfig, _ := config.LoadConfig()
	awsAliasProfile, _ := awsConfig.GetAliasProfile(Profile)

	s.Login(awsAliasProfile.SourceProfile)

	sessionName := fmt.Sprintf("xip-session-%d", time.Now().Unix())
	duration := int64(3600)

	input := sts.AssumeRoleInput{
		RoleArn:         &awsAliasProfile.RoleArn,
		ExternalId:      &awsAliasProfile.ExternalId,
		RoleSessionName: &sessionName,
		DurationSeconds: &duration,
	}

	output, err := options.Sts.AssumeRole(&input)
	if err != nil {
		panic(err)
	}

	fileName := cli.CreateRoleProfileFileName(awsAliasProfile.RoleArn, duration)
	ssoClientProfile := cli.NewRoleProfile(fileName, *output)
	ssoClientProfile.Save()
}

func (s *Sso) isAliasProfile(options *LoginOptions, Profile string) bool {
	_, err := options.Config.GetAliasProfile(Profile)

	return err == nil
}

func (s *Sso) hasValidSsoProfileWithAccessToken(Profile string) bool {
	awsConfig, _ := config.LoadConfig()
	awsSsoProfile, _ := awsConfig.GetSsoProfile(Profile)

	_, err := sso.LoadProfile(awsSsoProfile.StartUrl)

	return err == nil
}
