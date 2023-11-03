package app

import (
	"os"
	"time"
)

var (
	// AppName defines app name.
	AppName string

	// defines env name when running as dyno.
	Dyno string

	// Commit is the current commit of the application.
	Commit string

	// DeployedAt is the date of the current release.
	DeployedAt string

	// BuildInfo contains all build informations.
	BuildInfo map[string]any

	// BaseURL defines a base url to return as public endpoint.
	BaseURL string

	// EnableSitemap defines if sitemap is enabled.
	EnableSitemap bool

	// DefaultLocale is the default locale of the application.
	DefaultLocale = "en"
)

func init() {

	AppName = os.Getenv("APP_NAME")
	if AppName == "" {
		AppName = "prova"
	}

	Dyno = os.Getenv("DYNO")
	Commit = os.Getenv("GIT_REV")
	DeployedAt = time.Now().UTC().String()

	BaseURL = os.Getenv("BASE_URL")

	EnableSitemap = os.Getenv("ENABLE_SITEMAP") == "true"

	BuildInfo = map[string]any{
		"app_name":    AppName,
		"dyno":        Dyno,
		"commit":      Commit,
		"base_url":    BaseURL,
		"deployed_at": DeployedAt,
	}
}
