package cmd

import (
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	"github.com/spf13/cobra"
)

var auditKVCmd = &cobra.Command{
	Use:     "kv <collection> <primary key value>",
	Short:   "Audit your kv collection entry",
	Example: "immudb-log-audit audit kv samplecollection 100",
	Args:    cobra.MinimumNArgs(1),
	RunE:    auditKv,
}

func init() {
	auditCmd.AddCommand(auditKVCmd)
}

func auditKv(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	jr, err := immudb.NewJsonKVRepository(immuCli, args[0])
	if err != nil {
		return fmt.Errorf("could not create json kv repository, %w", err)
	}

	pkValue := ""
	if len(args) > 1 {
		pkValue = args[1]
	}

	history, err := jr.History(pkValue)
	if err != nil {
		return fmt.Errorf("could not get audit, %w", err)
	}

	for _, h := range history {
		fmt.Printf("{\"tx_id\": %d, \"revision\": %d, \"entry\": %s}\n", h.TxID, h.Revision, string(h.Entry))
	}

	return nil
}
