package cmd

import (
	"github.com/spf13/cobra"
)

func runParentCmdE(cmd *cobra.Command, args []string) error {
	if cmd.Parent() != nil && cmd.Parent().RunE != nil {
		err := cmd.Parent().RunE(cmd.Parent(), args)
		if err != nil {
			return err
		}
	}

	return nil
}
