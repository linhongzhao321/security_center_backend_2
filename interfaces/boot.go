package interfaces

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"infrastructure/aws"
	"infrastructure/contextextension"
	"infrastructure/logger"
	"interfaces/googleplay"
	"interfaces/kms"
	"interfaces/monitor"
	"interfaces/server"
	"interfaces/ssh"
)

func Root() *cobra.Command {
	var runMod contextextension.RunMod = `release`
	var logOutputPath, errorOutputPath string
	var awsAccessKey, awsAccessSecret, awsKmsRegion, awsKmsID string

	// PersistentPreRun 任何命令启动前的首要工作，如加载配置
	PersistentPreRun := func(cmd *cobra.Command, _ []string) error {
		// gen trace id
		cmd.SetContext(contextextension.GenTraceID(cmd.Context()))
		ctx := cmd.Context()

		// boot logger
		var options []logger.Option
		if contextextension.IsDebug(ctx) {
			options = append(options, logger.OutputMinLevel(zap.DebugLevel))
		} else {
			options = append(options, logger.OutputMinLevel(zap.InfoLevel))
		}
		if logOutputPath != `` {
			options = append(options, logger.WithOutputPaths(logOutputPath))
		}
		if errorOutputPath != `` {
			options = append(options, logger.WithErrorOutputPath(errorOutputPath))
		}

		err := logger.InitLogger(options...)
		if err != nil {
			return err
		}

		logger.Write(ctx, zap.InfoLevel, `boot.PersistentPreRun()...`)

		ctx = contextextension.WithMod(cmd.Context(), runMod)
		cmd.SetContext(ctx)
		logger.Write(ctx, zap.DebugLevel, `test debug output`)
		logger.Write(ctx, zap.InfoLevel, `running`, zap.Any(`run-mod`, runMod))

		// init kms
		if awsKmsRegion != `` && awsAccessKey != `` && awsAccessSecret != `` {
			err = aws.InitKMS(ctx, awsKmsRegion, awsAccessKey, awsAccessSecret, awsKmsID)
			if err != nil {
				logger.Write(ctx, zap.FatalLevel, `aws kms init error`, zap.Error(err))
				return err
			}
		}

		return nil
	}

	command := &cobra.Command{
		Use: `sec`,
		Short: `The security center backend service can be used to` +
			`start web backend services or execute some tools`,
		PersistentPreRunE: PersistentPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	command.PersistentFlags().
		VarP(&runMod, `run-mode`, `r`, fmt.Sprintf(`one of %v`, contextextension.EnabledRunMods))
	command.PersistentFlags().
		StringVar(&logOutputPath, `log-output`, ``, `log output path`)
	command.PersistentFlags().StringVar(&errorOutputPath, `error-output`, ``, `error log output path`)
	command.PersistentFlags().StringVar(&awsAccessKey, `aws-access-key`, ``, `key&secret with kms read-permissions`)
	command.PersistentFlags().StringVar(&awsAccessSecret, `aws-access-secret`, ``, `key&secret with kms read-permissions`)
	command.PersistentFlags().StringVar(&awsKmsRegion, `aws-kms-region`, `ap-east-1`, `kms instance region`)
	command.PersistentFlags().StringVar(&awsKmsID, `aws-kms-id`, ``, `kms key-id use for encrypt/decrypt`)
	command.MarkFlagsRequiredTogether(
		`aws-access-key`,
		`aws-access-secret`,
		`aws-kms-id`,
	)

	command.AddCommand(
		server.NewCommand(),
		kms.NewCommand(),
		monitor.NewCommand(),
		googleplay.NewCommand(),
		ssh.NewCommand(),
	)

	return command
}
