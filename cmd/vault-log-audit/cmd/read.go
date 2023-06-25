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

var readCmd = &cobra.Command{
	Use:   "read <collection> <<vault query>>",
	Short: "Read audit data from immudb key-value collection.",
	Example: `immudb-log-audit read samplecollection
immudb-log-audit read kv samplecollection indexed_field1=prefix1
immudb-log-audit read kv samplecollection indexed_field2=prefix2`,
	RunE: readKV,
}

func init() {
	rootCmd.AddCommand(readCmd)
}

func readKV(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	collection := "default"
	var query string
	if len(args) == 2 {
		collection = args[0]
		query = args[1]
	} else if len(args) == 1 {
		log.Info("Using default collection")
		query = args[0]
	}

	log.WithField("query", query).Debug("query")

	jsonRepository, err := vault.NewJsonVaultRepository(vaultClient, ledger, collection, flagBatchMode)
	if err != nil {
		return fmt.Errorf("could not initialize vault, %w", err)
	}

	jsons, err := jsonRepository.Read(query)
	if err != nil {
		return fmt.Errorf("could not read vault, %w", err)
	}

	for _, j := range jsons {
		fmt.Println(string(j))
	}

	return nil
}
