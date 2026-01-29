package googleplay

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"infrastructure/googleplay"
)

func Auth() *cobra.Command {
	outputFilename := ``
	code := ``

	command := &cobra.Command{
		Use:   `auth`,
		Short: `get googleplay oauth token from access code`,
		Long:  `please login for get oauth-token in cookie from accounts.google.com/embedded/setup/v2/android`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var token googleplay.Token
			err := token.Auth(code)
			if err != nil {
				return err
			}
			defer func() {
				if err != nil {
					fmt.Println(`write token file error, Please manually create and write the following content:`)
					fmt.Println(string(token.Data))
				}
			}()

			dir := filepath.Dir(outputFilename)
			if dir != `` && dir != `.` {
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					return err
				}
			}

			if outputFilename == `` {
				fmt.Println(`not specify output-filename , please record the following content yourself`)
				fmt.Println(string(token.Data))
				return nil
			}
			return os.WriteFile(outputFilename, token.Data, 0660)
		},
	}

	command.Flags().StringVarP(&outputFilename, `output`, `o`, outputFilename, ``)
	command.Flags().StringVar(&code, `code`, `c`, ``)
	err := command.MarkFlagRequired(`code`)
	if err != nil {
		panic(err)
	}

	return command
}
