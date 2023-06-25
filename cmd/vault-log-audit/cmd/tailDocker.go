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

var tailDockerCmd = &cobra.Command{
	Use:   "docker <collection> <container>",
	Short: "Tail from docker logs and store audit data in immudb vault collection. Collection needs to be created first.",
	Example: `immudb-log-audit tail docker default psql-postgresql-1 --follow --stdout --stderr
immudb-log-audit tail docker somecollection 3855fafd83b6 --stdout --stderr`,
	RunE: tailDocker,
	Args: cobra.MinimumNArgs(1),
}

func tailDocker(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	log.WithField("args", args).Info("Docker tail")

	lp, err := cmdutils.NewLineParser(flagParser)
	if err != nil {
		return fmt.Errorf("invalid line parser, %w", err)
	}

	collection := "default"
	var container string
	if len(args) == 2 {
		collection = args[0]
		container = args[1]
	} else {
		log.Info("Using default collection")
		container = args[0]
	}

	jsonRepository, err := vault.NewJsonVaultRepository(vaultClient, ledger, collection, flagBatchMode)
	if err != nil {
		return fmt.Errorf("could not initialize vault, %w", err)
	}

	flagSince, _ := cmd.Flags().GetString("since")
	flagStdout, _ := cmd.Flags().GetBool("stdout")
	flagStderr, _ := cmd.Flags().GetBool("stderr")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signals
		cancel()
	}()

	dockerTail, err := source.NewDockerTail(ctx, container, flagFollow, flagSince, flagStdout, flagStderr)
	if err != nil {
		return fmt.Errorf("invalide source: %w", err)
	}

	s := service.NewAuditService(dockerTail, lp, jsonRepository)
	err = s.Run()
	signal.Stop(signals)
	close(signals)
	return err
}

func init() {
	tailCmd.AddCommand(tailDockerCmd)
	tailDockerCmd.Flags().String("since", "", "since argument")
	tailDockerCmd.Flags().Bool("stdout", false, "If true, read stdout from container")
	tailDockerCmd.Flags().Bool("stderr", false, "If true, read stderr from container")
}
