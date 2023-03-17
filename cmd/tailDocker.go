package cmd

import (
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	"github.com/codenotary/immudb-log-audit/pkg/service"
	"github.com/codenotary/immudb-log-audit/pkg/source"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tailDockerCmd = &cobra.Command{
	Use:   "docker <collection> <container>",
	Short: "Tail from docker logs and store audit data in immudb collection. Collection needs to be created first.",
	Example: `immudb-log-audit tail docker pgaudit psql-postgresql-1 --follow --stdout --stderr
immudb-log-audit tail docker somecollection 3855fafd83b6 --stdout --stderr`,
	RunE: tailDocker,
	Args: cobra.ExactArgs(2),
}

func tailDocker(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	log.WithField("args", args).Info("Docker tail")

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

	flagSince, _ := cmd.Flags().GetString("since")
	flagStdout, _ := cmd.Flags().GetBool("stdout")
	flagStderr, _ := cmd.Flags().GetBool("stderr")
	dockerTail, err := source.NewDockerTail(args[1], flagFollow, flagSince, flagStdout, flagStderr)
	if err != nil {
		return fmt.Errorf("invalide source: %w", err)
	}

	s := service.NewAuditService(dockerTail, lp, jsonRepository)
	return s.Run()
}

func init() {
	tailCmd.AddCommand(tailDockerCmd)
	tailDockerCmd.Flags().String("since", "", "since argument")
	tailDockerCmd.Flags().Bool("stdout", false, "If true, read stdout from container")
	tailDockerCmd.Flags().Bool("stderr", false, "If true, read stderr from container")
}
