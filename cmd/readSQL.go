package cmd

import (
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	"github.com/spf13/cobra"
)

var readSQLCmd = &cobra.Command{
	Use:   "sql <collection> <<SQL query conditions>>",
	Short: "Read audit data from immudb SQL collection.",
	RunE:  readSQL,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	readCmd.AddCommand(readSQLCmd)
}

func readSQL(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	jr, err := immudb.NewJsonSQLRepository(immuCli, args[0])
	if err != nil {
		return fmt.Errorf("could not create json kv repository, %w", err)
	}

	query := ""
	if len(args) == 2 {
		query = args[1]
	}

	jsons, err := jr.Read(query)
	if err != nil {
		return fmt.Errorf("could not read, %w", err)
	}

	for _, j := range jsons {
		fmt.Println(string(j))
	}

	return nil
}
