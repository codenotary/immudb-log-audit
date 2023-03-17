package cmd

import (
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit data from immudb",
	RunE:  audit,
}

func init() {
	rootCmd.AddCommand(auditCmd)
}

func audit(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "audit" {
		return cmd.Help()
	}

	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	return nil
}
