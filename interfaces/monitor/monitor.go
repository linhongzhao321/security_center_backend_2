package monitor

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   `monitor`,
		Short: `run all monitor, or run specified monitor through sub-commands`,
	}

	command.AddCommand(NewReleaseConsistency())
	return command
}
