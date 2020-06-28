package sso

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	ssos "github.com/aws/aws-sdk-go/service/sso"
	"github.com/aws/aws-sdk-go/service/ssooidc"
	"time"
	"xip/aws/functions/config/app"
	"xip/aws/functions/config/config"
	"xip/aws/functions/config/credentials"
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
	_AwsSession    *session.Session
	_SsoClient     *ssos.SSO
	_SsoOidcClient *ssooidc.SSOOIDC

	// Configuration clients
	AppConfig         *app.App
	ConfigConfig      *config.Config
	CredentialsConfig *credentials.Credentials

	// Persisted information
	ClientId         *string
	ClientSecret     *string
	ClientExpiration *time.Time

	// On-the-fly information
	DeviceCode           *string
	DeviceCodeExpiration *int32
	UserCode             *string
	AccessToken          *string
}

// https://docs.aws.amazon.com/cognito/latest/developerguide/token-endpoint.html
// https://docs.aws.amazon.com/singlesignon/latest/OIDCAPIReference/API_CreateToken.html
// https://docs.aws.amazon.com/singlesignon/latest/PortalAPIReference/API_GetRoleCredentials.html

func New(awsSession session.Session, appConfig app.App, awsConfig config.Config) Sso {
	// App config values
	appConfigValues := appConfig.Get()

	// Creds instance
	creds := credentials.New(*appConfigValues.AwsConfigPath, *appConfigValues.DefaultProfile)

	sso := Sso{
		_AwsSession: &awsSession,

		AppConfig:         &appConfig,
		ConfigConfig:      &awsConfig,
		CredentialsConfig: &creds,
	}
	sso._Setup()
	sso._LoadConfig()

	return sso
}

func (sso *Sso) _Setup() {
	// Create a SSOOIDC app with additional configuration
	sso._SsoClient = ssos.New(sso._AwsSession)
	sso._SsoOidcClient = ssooidc.New(sso._AwsSession)

}

func (sso *Sso) Login() {
	// Register the device if needed
	sso._RegisterClient()

	// Authorize the device
	sso._AuthorizeDevice()

	// Retrieve the role credentials by assuming it
	sso._RetrieveRoleCredentials()
}

func (sso *Sso) Configure(values ConfigureValues) {
	// Update defalt profile
	appValues := sso.AppConfig.Get()
	appValues.DefaultProfile = values.Profile
	sso.AppConfig.Set(appValues)

	// Save the new configuration
	ConfigConfig := config.New(*sso.AppConfig)
	ConfigConfig.SetSsoProfile(config.SsoProfile{
		StartUrl:  *values.StartUrl,
		Region:    *values.Region,
		AccountId: *values.AccountId,
		Role:      *values.RoleName,
		Output:    "json",
	})
	sso.ConfigConfig = &ConfigConfig

	// Set region
	sso._AwsSession.Config.Region = values.Region
	sso._Setup()

	// Register the device if needed
	sso.Login()
}

func (sso *Sso) _LoadConfig() {
	if !sso.ConfigConfig.Valid() {
		return
	}

	clientValues := sso.AppConfig.Get()

	sso.ClientId = clientValues.ClientId
	sso.ClientSecret = clientValues.ClientSecret
	sso.ClientExpiration = clientValues.ClientExpiration
}

func (sso *Sso) _SaveConfig() {
	values := sso.AppConfig.Get()

	values.ClientId = sso.ClientId
	values.ClientSecret = sso.ClientSecret
	values.ClientExpiration = sso.ClientExpiration

	sso.AppConfig.Set(values)
}

func (sso *Sso) _RegisterClient() {
	if sso.AppConfig.Valid() {
		return
	}

	clientName := "xip_cli_tool"
	clientType := "public"

	clientInput := &ssooidc.RegisterClientInput{
		ClientName: &clientName,
		ClientType: &clientType,
		Scopes:     nil,
	}

	output, err := sso._SsoOidcClient.RegisterClient(clientInput)
	if err != nil {
		panic(err)
	}

	expiration := helpers.IntToTime(int(*output.ClientSecretExpiresAt))

	sso.ClientId = output.ClientId
	sso.ClientSecret = output.ClientSecret
	sso.ClientExpiration = &expiration

	sso._SaveConfig()
}

func (sso *Sso) _AuthorizeDevice() {
	if sso.DeviceCodeExpiration != nil && *sso.DeviceCodeExpiration > int32(time.Now().Unix()) {
		return
	}

	conf := sso.ConfigConfig.GetSsoProfile()

	clientInput := &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     sso.ClientId,
		ClientSecret: sso.ClientSecret,
		StartUrl:     &conf.StartUrl,
	}

	output, err := sso._SsoOidcClient.StartDeviceAuthorization(clientInput)
	if err != nil {
		panic(err)
	}

	helpers.OpenBrowser(*output.VerificationUriComplete)

	tokenExpiration := int32(time.Now().Unix()) + int32(*output.ExpiresIn)

	sso.UserCode = output.UserCode
	sso.DeviceCode = output.DeviceCode
	sso.DeviceCodeExpiration = &tokenExpiration

	sso._CreateToken(int(*output.ExpiresIn)/int(*output.Interval), int(*output.Interval))
}

func (sso *Sso) _CreateToken(retryCount int, interval int) {
	grantType := "urn:ietf:params:oauth:grant-type:device_code"

	clientInput := &ssooidc.CreateTokenInput{
		ClientId:     sso.ClientId,
		ClientSecret: sso.ClientSecret,
		DeviceCode:   sso.DeviceCode,
		Code:         sso.UserCode,
		GrantType:    &grantType,
	}

	sleepTimeout, _ := time.ParseDuration(fmt.Sprintf("%ds", interval))

	for i := 0; i < retryCount; i++ {
		output, err := sso._SsoOidcClient.CreateToken(clientInput)
		if err != nil {
			time.Sleep(sleepTimeout)

			continue
		}

		sso.AccessToken = output.AccessToken

		return
	}

	panic("Failed to create token")
}

func (sso *Sso) _RetrieveRoleCredentials() {
	if sso.CredentialsConfig.Valid() {
		return
	}

	conf := sso.ConfigConfig.GetSsoProfile()

	input := ssos.GetRoleCredentialsInput{
		AccessToken: sso.AccessToken,
		AccountId:   &conf.AccountId,
		RoleName:    &conf.Role,
	}

	output, err := sso._SsoClient.GetRoleCredentials(&input)
	if err != nil {
		panic(err)
	}

	sso.CredentialsConfig.FromRoleCredentials(conf.Region, *output.RoleCredentials)
}
