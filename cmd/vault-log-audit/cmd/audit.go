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
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/repository/vault"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:     "audit <collection> <documentID>",
	Short:   "Audit your collection entry",
	Example: "vault-log-audit audit default 648a32500000000000000bc0b922a7c9",
	Args:    cobra.MinimumNArgs(1),
	RunE:    audit,
}

func init() {
	rootCmd.AddCommand(auditCmd)
}

func audit(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	collection := "default"
	var documentID string
	if len(args) == 2 {
		collection = args[0]
		documentID = args[1]
	} else {
		log.Info("Using default collection")
		documentID = args[0]
	}

	jsonRepository, err := vault.NewJsonVaultRepository(vaultClient, ledger, collection, flagBatchMode)
	if err != nil {
		return fmt.Errorf("could not initialize vault, %w", err)
	}

	history, err := jsonRepository.Audit(documentID)
	if err != nil {
		return fmt.Errorf("could not audit, %w", err)
	}

	for _, h := range history {
		fmt.Println(string(h))
	}

	return nil
}
