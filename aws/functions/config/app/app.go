package app

import (
	"time"
	"xip/utils/config_file/ini"
)

type App struct {
	File *ini.ConfigFileIni
}

type Values struct {
	DefaultProfile   *string
	ClientId         *string
	ClientSecret     *string
	ClientExpiration *time.Time
	AwsConfigPath    *string
}

func New() App {
	return App{
		File: ini.AppConf(),
	}
}

func (app *App) Set(input Values) {
	var expiration *string
	if input.ClientExpiration != nil {
		_expiration := input.ClientExpiration.Local().Format(time.RFC3339)
		expiration = &_expiration
	}

	_ = app.File.Read()
	app.File.Set("aws.default_profile", input.DefaultProfile)
	app.File.Set("aws.client_id", input.ClientId)
	app.File.Set("aws.client_secret", input.ClientSecret)
	app.File.Set("aws.client_expiration", expiration)
	app.File.Set("aws.config_path", input.AwsConfigPath)
	_ = app.File.Write()
}

func (app *App) Get() Values {
	_ = app.File.Read()

	var (
		timeParsed *time.Time
		timeRaw    *string = app.File.GetStringOptional("aws.client_expiration")
	)

	if timeRaw != nil {
		_timeParsed, _ := time.Parse(time.RFC3339, *timeRaw)
		timeParsed = &_timeParsed
	}

	return Values{
		DefaultProfile:   app.File.GetStringOptional("aws.default_profile"),
		ClientId:         app.File.GetStringOptional("aws.client_id"),
		ClientSecret:     app.File.GetStringOptional("aws.client_secret"),
		ClientExpiration: timeParsed,
		AwsConfigPath:    app.File.GetStringOptional("aws.config_path"),
	}
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
