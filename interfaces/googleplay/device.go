package googleplay

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"infrastructure/googleplay"
	"infrastructure/logger"
)

func Device() *cobra.Command {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	currentPath := filepath.Dir(ex)
	outputFilename := currentPath + string(os.PathSeparator) + `device.bin`
	command := &cobra.Command{
		Use:   `device`,
		Short: `building a bin file for simulating Android`,
		RunE: func(cmd *cobra.Command, args []string) error {
			googleplay.Phone.ABI = googleplay.ABIs[2]
			var checkin googleplay.Checkin
			if err = checkin.Checkin(googleplay.Phone); err != nil {
				return err
			}
			logger.Write(cmd.Context(), zap.InfoLevel,
				`device info wrote to file`,
				zap.String(`filename`, outputFilename),
			)
			if err = os.WriteFile(outputFilename, checkin.Data, 0666); err != nil {
				return err
			}

			// check device-info validity
			if err = checkin.Unmarshal(); err != nil {
				return err
			}
			err = checkin.Sync(googleplay.Phone)
			return err
		},
	}
	command.Flags().StringVarP(&outputFilename, `output`, `o`, outputFilename, ``)

	return command
}
