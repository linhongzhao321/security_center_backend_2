package ssh

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   `ssh`,
		Short: `ssh tools, such as generate ssh keypair, generate ssh ca`,
		//PreRunE: func(cmd *cobra.Command, args []string) error {
		//	return nil
		//},
	}
	command.AddCommand(GenKeypair())
	command.AddCommand(GithubCA())
	return command
}
