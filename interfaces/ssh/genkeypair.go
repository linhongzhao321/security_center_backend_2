package ssh

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"infrastructure/logger"
	"infrastructure/sshkeygen"
)

func GenKeypair() *cobra.Command {
	command := &cobra.Command{
		Use:   `gen-keypair`,
		Short: `ssh tools, such as generate ssh keypair, generate ssh ca`,
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKey, publicKey, err := sshkeygen.GenerateKeypair()
			if err != nil {
				logger.Write(cmd.Context(), zap.ErrorLevel, `generate keypair error`, zap.Error(err))
				return err
			}
			fmt.Println(privateKey)
			fmt.Println(publicKey)
			return nil
		},
	}
	return command
}
