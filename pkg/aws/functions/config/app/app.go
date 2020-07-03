package app

import (
	"xip/utils/config_file/ini"
)

type App struct {
	File *ini.ConfigFileIni
}

type Values struct {
	DefaultProfile *string
}

func New() App {
	return App{
		File: ini.AppConf(),
	}
}

func (app *App) Set(input Values) {
	_ = app.File.Read()
	app.File.Set("aws.default_profile", input.DefaultProfile)
	_ = app.File.Write()
}

func (app *App) Get() Values {
	_ = app.File.Read()

	return Values{
		DefaultProfile: app.File.GetStringOptional("aws.default_profile"),
	}
}
