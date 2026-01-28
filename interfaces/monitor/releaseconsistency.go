package monitor

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/golang-jwt/jwt/v5"
	errors2 "github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"infrastructure/aws"
	"infrastructure/github"
	"infrastructure/googleplay"
	"infrastructure/ios"
	"infrastructure/logger"
	"infrastructure/webpage"
	"internal/core"
	"internal/monitor"
	"internal/project"
)

var config *Config

func NewReleaseConsistency() *cobra.Command {
	command := &cobra.Command{
		Use:   `release-consistency`,
		Short: `check release consistency`,
		Long: `
check release consistency between git latest release and production.
before run:
1. applying for KMS-KEY-ID from AWS
2. run "sec secret"，enter the secret according to the instructions to obtain the encrypted secret
3. set environment
     AWS_ACCESS_KEY
     AWS_ACCESS_SECRET
     GITHUB_API_TOKEN (encrypted by kms)
     SLACK_WEBHOOK_URL (encrypted by kms)
`,
		Example: `release-consistency ` +
			`-aws-access-key <string> ` +
			`-aws-access-secret <secret> ` +
			`-slack-webhook <webhook url> ` +
			`-slack-id <id-1> <id-2> <id-3>` +
			`-t <time string>`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger.Write(ctx, zap.InfoLevel, `boot.PersistentPreRun()...`)
			var err error
			config, err = NewConfig()
			if err != nil {
				return err
			}
			if aws.IsInitialized() {
				err := aws.DecryptStrings(&config.SlackWebhook, &config.IpatoolPassphrase, &config.PagerDutyEventRoutingKey)
				if err != nil {
					logger.Write(ctx, zap.FatalLevel, `aws decode fail`, zap.Error(err))
					return err
				}
			}

			// 设置 default http transport
			// 较多地方有用到 http.DefaultClient，改造成本较高
			// 因此，直接修改 http.DefaultClient.Transport
			// 确保链接复用资源充足即可
			netDialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			http.DefaultClient.Transport = &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           netDialer.DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				MaxConnsPerHost:       0,
				IdleConnTimeout:       5 * time.Minute,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}

			return nil
		},
		RunE: RunE,
	}

	return command
}

func RunE(cmd *cobra.Command, _ []string) error {

	monitorNotifiers := core.MonitorNotifiers{}
	if config.SlackWebhook != `` {
		monitorNotifiers = append(monitorNotifiers, monitor.SlackNotifier(config.SlackWebhook, config.SlackIDs...))
	}
	if config.PagerDutyEventRoutingKey != `` {
		monitorNotifiers = append(monitorNotifiers, monitor.PagerDutyNotifier(config.PagerDutyEventRoutingKey))
	}
	if config.PeriodicReportSlack != `` {
		monitorNotifiers = append(monitorNotifiers, monitor.PeriodicReportSlack(config.SlackWebhook, config.PeriodicReportSlack))
	}

	ctx := cmd.Context()

	if len(config.Projects) == 0 {
		logger.Write(ctx, zap.InfoLevel, `project config is empty, exit.`)
		return nil
	}
	projectInstances, err := loadProjects(ctx, monitorNotifiers)
	if err != nil {
		return err
	}

	// run once or run for time schedule
	if config.TTL == `` {
		runner(cmd.Context(), projectInstances...)
	} else {
		timezone, _ := time.LoadLocation("Asia/Shanghai")
		s := gocron.NewScheduler(timezone)
		_, err := s.Every(config.TTL).Do(func() {
			runner(cmd.Context(), projectInstances...)
		})
		if err != nil {
			logger.Write(ctx, zap.ErrorLevel, `s.Every(ttl).Do()`, zap.Error(err))
			return err
		}
		s.StartBlocking()
	}

	return nil
}

func loadProjects(ctx context.Context, monitorNotifiers core.MonitorNotifiers) ([]core.Project, error) {
	var projectInstances []core.Project
	for _, projectConfig := range config.Projects {
		logger.Write(ctx, zap.InfoLevel, `load project`, zap.String(`name`, projectConfig.Name))
		var options []project.Option
		if projectConfig.IsSync {
			options = append(options, project.SyncRunMonitor())
		}
		if len(monitorNotifiers) > 0 {
			options = append(options, project.AppendMonitors(
				monitor.NewReleaseConsistencyMonitor(monitorNotifiers...),
			))
		}
		for _, component := range projectConfig.Components {
			githubAppPEM, err := os.ReadFile(component.Github.GithubAppPEMFilename)
			if err != nil {
				logger.Write(ctx, zap.FatalLevel, `get github app pem fail`, zap.Error(err))
				return nil, err
			}
			privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(githubAppPEM)
			if err != nil {
				logger.Write(context.Background(), zap.FatalLevel, `ParseRSAPrivateKeyFromPEM() error`, zap.Error(err))
				return nil, errors2.Wrap(err, `ParseRSAPrivateKeyFromPEM() error`)
			}
			if component.Type == `android` {
				withChannels := []project.ComponentOption{
					project.WithVersionControlService(
						func(ctx context.Context, latest uint8) (releases core.Releases, err error) {
							client, err := github.GetClientByApp(component.Github.GithubAppClientID, privateKey, component.Github.InstallationID)
							if err != nil {
								return nil, err
							}
							return client.GetAppReleaseFromAssets(ctx, component.Github.Owner, component.Github.Repo, ``, int(latest), core.ReleaseTypeAPK)
						},
					),
				}
				for _, releasesChannels := range component.Releases {
					if releasesChannels.Channel == `googleplay` {
						if component.IsAppBundle {
							withChannels = append(withChannels, project.WithReleaseChannel(
								googleplay.MustNewOAuthReleaseChannel(
									component.BundleID,
									config.GooglePlayOAuthCode,
									config.GooglePlayCredentialFilename,
								),
							))
						} else {
							withChannels = append(withChannels, project.WithReleaseChannel(
								googleplay.MustNewReleaseChannel(config.GooglePlayTokenFilename, config.GooglePlayDeviceFilename, component.BundleID),
							))
						}
					} else if releasesChannels.Channel == `webpage` {
						if len(releasesChannels.URL) == 0 {
							logger.Write(ctx, zap.FatalLevel, `releases.*.url is empty`,
								zap.String(`project`, projectConfig.Name),
								zap.String(`component`, component.Type),
							)
						}
						var options []webpage.Option
						if releasesChannels.IsDownloadURL {
							options = append(options, webpage.DownloadURL(releasesChannels.URL))
						} else {
							options = append(options, webpage.FindURLByRegExp(releasesChannels.URL, webpage.RegexpApkURI))
						}
						withChannels = append(withChannels,
							project.WithReleaseChannel(
								webpage.MustReleaseChannelForAPK(
									options...,
								),
							),
						)
					}
				}
				options = append(options, project.MustComponent(component.Type, withChannels...))
			} else if component.Type == `ios` {
				options = append(options,
					project.MustComponent(`ios`,
						project.WithReleaseChannel(ios.NewAppStore(ctx, config.IpatoolPassphrase, component.BundleID, component.BinaryFilename)),
						project.WithVersionControlService(func(ctx context.Context, latest uint8) (release core.Releases, err error) {
							client, err := github.GetClientByApp(component.Github.GithubAppClientID, privateKey, component.Github.InstallationID)
							if err != nil {
								return nil, err
							}
							return client.GetAppReleaseFromAssets(ctx, component.Github.Owner, component.Github.Repo, component.BinaryFilename, int(latest), core.ReleaseTypeIPA)
						}),
					),
				)
			} else if component.Type == `frontend` {
				options = append(options,
					project.MustComponent(`frontend`,
						project.WithReleaseChannel(
							webpage.MustNewReleaseChannel(component.Releases[0].URL),
						),
						project.WithVersionControlService(
							func(ctx context.Context, _ uint8) (release core.Releases, err error) {
								client, err := github.GetClientByApp(component.Github.GithubAppClientID, privateKey, component.Github.InstallationID)
								if err != nil {
									return nil, err
								}
								return client.GetWebReleases(ctx, component.Github.Owner, component.Github.Repo, component.CorePaths...)
							},
						),
					),
				)
			} else {
				logger.Write(ctx, zap.ErrorLevel, `undiended component type`,
					zap.String(`project`, projectConfig.Name),
					zap.String(`component`, component.Type),
				)
			}
		}
		projectInstance := project.MustProject(projectConfig.Name, options...)
		projectInstances = append(projectInstances, projectInstance)
		logger.Write(ctx, zap.InfoLevel, `load project ended`, zap.String(`name`, projectConfig.Name))
	}
	return projectInstances, nil
}

func runner(ctx context.Context, projectInstances ...core.Project) {
	for _, projectInstance := range projectInstances {
		logger.Write(ctx, zap.InfoLevel, `project start`, zap.String(`name`, projectInstance.Name()))
		results := projectInstance.Run(ctx)
		for _, result := range results {
			if result.IsFail() {
				logger.Write(ctx, zap.WarnLevel, result.String(), zap.Error(result.Error()))
			} else {
				logger.Write(ctx, zap.InfoLevel, result.String())
			}
		}
		logger.Write(ctx, zap.InfoLevel, `project ended`, zap.String(`name`, projectInstance.Name()))
	}
}
