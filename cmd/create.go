package cmd

import (
	"github.com/spf13/cobra"
)

var flagParser string
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create collection in immudb",
	RunE:  create,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().StringVar(&flagParser, "parser", "", "Line parser to be used. When not specified, lines will be considered as jsons. Also available 'pgaudit', 'wrap'. For those, indexes are predefined.")
}

func create(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "create" {
		return cmd.Help()
	}

	return runParentCmdE(cmd, args)
}
