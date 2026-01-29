package monitor

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"infrastructure/logger"
)

type Config struct {
	TTL                          string   `mapstruct:"ttl"`
	IpatoolPassphrase            string   `mapstruct:"ipatoolPassphrase"`
	GooglePlayOAuthCode          string   `mapstruct:"googlePlayOAuthCode"`
	GooglePlayCredentialFilename string   `mapstruct:"googlePlayCredentialFilename"`
	GooglePlayTokenFilename      string   `mapstruct:"googlePlayTokenFilename"`
	GooglePlayDeviceFilename     string   `mapstruct:"googlePlayDeviceFilename"`
	SlackWebhook                 string   `mapstruct:"slackWebhook"`
	SlackIDs                     []string `mapstruct:"slackIDs"`
	PagerDutyEventRoutingKey     string   `mapstruct:"pagerDutyEventRoutingKey"`
	PeriodicReportSlack          string   `mapstruct:"periodicReportSlack"`
	Projects                     []struct {
		Name       string `mapstruct:"name"`
		IsSync     bool   `mapstruct:"isSync"`
		Components []struct {
			Type        string `mapstruct:"type"`
			BundleID    string `mapstruct:"bundleID,omitempty"`
			IsAppBundle bool   `mapstruct:"IsAppBundle,omitempty"`
			Github      struct {
				InstallationID       int64  `mapstruct:"installationID"`
				Owner                string `mapstruct:"owner"`
				Repo                 string `mapstruct:"repo"`
				GithubAppClientID    string `mapstruct:"githubAppClientID"`
				GithubAppPEMFilename string `mapstruct:"githubAppPEMFilename"`
			} `mapstruct:"github"`
			Releases []struct {
				Channel       string `mapstruct:"channel"`
				URL           string `mapstruct:"url"`
				IsDownloadURL bool   `mapstruct:"isDownloadURL"`
			} `mapstruct:"releases,omitempty"`
			CorePaths      []string `mapstruct:"corePaths,omitempty"`
			BinaryFilename string   `mapstruct:"binaryFilename,omitempty"`
		} `mapstruct:"components"`
	} `mapstruct:"projects"`
}

const configPath string = `resources/config/monitor`

func NewConfig() (*Config, error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigType(`json`)
	err := viper.ReadInConfig()
	if err != nil {
		logger.Write(context.Background(), zap.FatalLevel, `read config fail`, zap.Error(err))
		return nil, err
	}
	config := &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		logger.Write(context.Background(), zap.FatalLevel, `read unmarshal fail`, zap.Error(err))
		return nil, err
	}
	return config, nil
}
