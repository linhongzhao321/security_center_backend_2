package googleplay

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   `google-play`,
		Short: `tools about google-play`,
	}
	command.AddCommand(
		Auth(),
		Device(),
		Download(),
	)
	return command
}
