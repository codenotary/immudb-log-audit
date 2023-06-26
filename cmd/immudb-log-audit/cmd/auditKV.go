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
