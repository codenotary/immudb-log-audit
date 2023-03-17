package cmd

import (
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/lineparser"
	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	"github.com/codenotary/immudb-log-audit/pkg/service"
	"github.com/spf13/cobra"
)

var flagFollow bool

var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Tail your source and store audit data in immudb",
	RunE:  tail,
}

func init() {
	rootCmd.AddCommand(tailCmd)
	tailCmd.PersistentFlags().BoolVar(&flagFollow, "follow", false, "If True, follow data stream. The follower supports file rotation.")
}

func tail(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "tail" {
		return cmd.Help()
	}

	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

func newLineParser(name string) (service.LineParser, error) {
	var lp service.LineParser
	switch name {
	case "":
		lp = lineparser.NewDefaultLineParser()
	case "pgaudit":
		lp = lineparser.NewPGAuditLineParser()
	case "wrap":
		lp = lineparser.NewWrapLineParser()
	default:
		return nil, fmt.Errorf("not supported parser: %s", flagParser)
	}

	return lp, nil
}

func newJsonRepository(rType string, collection string) (service.JsonRepository, error) {
	var jsonRepository service.JsonRepository
	var err error
	switch rType {
	case "kv":
		jsonRepository, err = immudb.NewJsonKVRepository(immuCli, collection)
		if err != nil {
			return nil, fmt.Errorf("could not create json repository, %w", err)
		}
	case "sql":
		jsonRepository, err = immudb.NewJsonSQLRepository(immuCli, collection)
		if err != nil {
			return nil, fmt.Errorf("could not create json repository, %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid repository type %s", rType)
	}
	return jsonRepository, nil
}
