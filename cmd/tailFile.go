package cmd

import (
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	"github.com/codenotary/immudb-log-audit/pkg/service"
	"github.com/codenotary/immudb-log-audit/pkg/source"
	"github.com/spf13/cobra"
)

var tailFileCmd = &cobra.Command{
	Use:   "file <collection> <file>",
	Short: "Tail from file and store audit data in immudb collection.",
	Example: `immudb-log-audit tail file k8scollection kubernetes.log --follow
immudb-log-audit tail file somecollection /path/to/log/file`,
	RunE: tailFile,
	Args: cobra.ExactArgs(2),
}

func tailFile(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	cfg, err := immudb.NewConfigs(immuCli).Read(args[0])
	if err != nil {
		return fmt.Errorf("collection does not exist, please create one first, %w", err)
	}

	lp, err := newLineParser(cfg.Parser)
	if err != nil {
		return fmt.Errorf("collection configuration is corrupted, %w", err)
	}

	jsonRepository, err := newJsonRepository(cfg.Type, args[0])
	if err != nil {
		return fmt.Errorf("collection configuration is corrupted, %w", err)
	}

	fileTail, err := source.NewFileTail(args[1], flagFollow)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	s := service.NewAuditService(fileTail, lp, jsonRepository)
	return s.Run()
}

func init() {
	tailCmd.AddCommand(tailFileCmd)
}
