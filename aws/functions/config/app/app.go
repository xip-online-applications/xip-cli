package app

import (
	"time"
	"xip/utils/config_file/ini"
)

type App struct {
	File *ini.ConfigFileIni
}

type Values struct {
	DefaultProfile        *string
	AwsConfigPath         *string
	ClientId              *string
	ClientSecret          *string
	ClientExpiration      *time.Time
	AccessToken           *string
	AccessTokenExpiration *time.Time
}

func New() App {
	return App{
		File: ini.AppConf(),
	}
}

func (app *App) Set(input Values) {
	var clientExpiration *string
	if input.ClientExpiration != nil {
		_expiration := input.ClientExpiration.Local().Format(time.RFC3339)
		clientExpiration = &_expiration
	}

	var accessTokenExpiration *string
	if input.AccessTokenExpiration != nil {
		_expiration := input.AccessTokenExpiration.Local().Format(time.RFC3339)
		accessTokenExpiration = &_expiration
	}

	_ = app.File.Read()
	app.File.Set("aws.default_profile", input.DefaultProfile)
	app.File.Set("aws.config_path", input.AwsConfigPath)
	app.File.Set("aws.client_id", input.ClientId)
	app.File.Set("aws.client_secret", input.ClientSecret)
	app.File.Set("aws.client_expiration", clientExpiration)
	app.File.Set("aws.access_token", input.AccessToken)
	app.File.Set("aws.access_token_expiration", accessTokenExpiration)
	_ = app.File.Write()
}

func (app *App) Get() Values {
	_ = app.File.Read()

	return Values{
		DefaultProfile:        app.File.GetStringOptional("aws.default_profile"),
		AwsConfigPath:         app.File.GetStringOptional("aws.config_path"),
		ClientId:              app.File.GetStringOptional("aws.client_id"),
		ClientSecret:          app.File.GetStringOptional("aws.client_secret"),
		ClientExpiration:      app.ParseToTime(app.File.GetStringOptional("aws.client_expiration")),
		AccessToken:           app.File.GetStringOptional("aws.access_token"),
		AccessTokenExpiration: app.ParseToTime(app.File.GetStringOptional("aws.access_token_expiration")),
	}
}

func (app *App) ParseToTime(timeRaw *string) *time.Time {
	var (
		timeParsed *time.Time
	)

	if timeRaw != nil {
		_timeParsed, _ := time.Parse(time.RFC3339, *timeRaw)
		timeParsed = &_timeParsed
	}

	return timeParsed
}

func (app *App) Initialized() bool {
	values := app.Get()

	return values.AwsConfigPath != nil
}

func (app *App) Valid() bool {
	values := app.Get()

	if values.AwsConfigPath == nil {
		return false
	}

	return values.ClientExpiration == nil || values.ClientExpiration.After(time.Now())
}
