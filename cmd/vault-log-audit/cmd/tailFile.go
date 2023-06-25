/*
Copyright 2023 Codenotary Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cmdutils "github.com/codenotary/immudb-log-audit/pkg/cmd"
	"github.com/codenotary/immudb-log-audit/pkg/repository/vault"
	"github.com/codenotary/immudb-log-audit/pkg/service"
	"github.com/codenotary/immudb-log-audit/pkg/source"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagRegistryEnabled bool
	flagRegistryDBDir   string
)

var tailFileCmd = &cobra.Command{
	Use:   "file <collection> <file>",
	Short: "Tail from file and store audit data in immudb collection.",
	Example: `immudb-log-audit tail file k8scollection kubernetes.log --follow
immudb-log-audit tail file somecollection /path/to/log/file`,
	RunE: tailFile,
	Args: cobra.MinimumNArgs(1),
}

func tailFile(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	lp, err := cmdutils.NewLineParser(flagParser)
	if err != nil {
		return fmt.Errorf("invalid line parser, %w", err)
	}

	collection := "default"
	var file string
	log.WithField("args", args).Debug("Args")
	if len(args) == 2 {
		collection = args[0]
		file = args[1]
	} else {
		log.Info("Using default collection")
		file = args[0]
	}

	jsonRepository, err := vault.NewJsonVaultRepository(vaultClient, ledger, collection, flagBatchMode)
	if err != nil {
		return fmt.Errorf("could not initialize vault, %w", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signals
		cancel()
	}()

	fileTail, err := source.NewFileTail(ctx, file, flagFollow, flagRegistryEnabled, flagRegistryDBDir)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	s := service.NewAuditService(fileTail, lp, jsonRepository)

	err = s.Run()
	signal.Stop(signals)
	close(signals)
	return err
}

func init() {
	tailCmd.AddCommand(tailFileCmd)
	tailFileCmd.Flags().BoolVar(&flagRegistryEnabled, "file-registry-enabled", true, "Enable monitoring of read files")
	tailFileCmd.Flags().StringVar(&flagRegistryDBDir, "file-registry-dir", "", "Directory where registry of monitored files should be stored, default is current directory")
}
