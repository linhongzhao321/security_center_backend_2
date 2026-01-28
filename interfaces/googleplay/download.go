package googleplay

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"infrastructure/googleplay"
	"infrastructure/logger"
)

const permTemp = 0660

func Download() *cobra.Command {
	var bundleID, tokenFilename, deviceFilename, outputFilename string
	command := &cobra.Command{
		Use:   `download`,
		Short: `download latest version for apk using bundle id`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := googleplay.NewDeviceClient(tokenFilename, deviceFilename)
			bs, err := client.DownloadAPK(cmd.Context(), bundleID)
			if err != nil {
				return err
			}
			if outputFilename == `` {
				outputFilename = strings.Replace(bundleID, `.`, `_`, 1) + `.apk`
			}
			f, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_CREATE, permTemp)
			if err != nil {
				return err
			}
			_, err = f.Write(bs)
			if err != nil {
				return err
			}
			logger.Write(cmd.Context(), zap.InfoLevel, `apk file generated`, zap.String(`filename`, outputFilename))
			return f.Close()
		},
	}
	command.Flags().StringVarP(&tokenFilename, `token`, `t`, ``, `token-filename, see sub-command "google-play auth -h"`)
	command.Flags().StringVarP(&deviceFilename, `device`, `d`, `./resources/device.bin`,
		`device-filename, see sub-command "google-play device -h"`,
	)
	command.Flags().StringVarP(&bundleID, `bundle-id`, `b`, ``, ``)
	command.Flags().StringVarP(&outputFilename, `output`, `o`, ``, `apk filename, default is ./<bundle-id>.apk`)
	command.MarkFlagsRequiredTogether(`bundle-id`, `token`)

	return command
}
