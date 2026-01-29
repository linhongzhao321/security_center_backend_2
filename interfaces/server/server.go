package server

import (
	"github.com/spf13/cobra"

	_ "interfaces/server/controllers"
	"interfaces/server/router"
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   `server`,
		Short: `run backend-server`,
		Long:  `run backend-server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return router.Run()
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return command
}
