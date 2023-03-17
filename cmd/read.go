package cmd

import (
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read audit data from immudb.",
	RunE:  read,
}

func init() {
	rootCmd.AddCommand(readCmd)
}

func read(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "read" {
		return cmd.Help()
	}

	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	return nil
}
