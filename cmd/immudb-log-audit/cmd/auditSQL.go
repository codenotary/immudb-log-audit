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

var auditSQLCmd = &cobra.Command{
	Use:     "sql <collection> <<temporal query range and condition>>",
	Short:   "Audit your sql collection with temporal queries",
	Example: "immudb-log-audit audit sql samplecollection \"SINCE '2022-01-06 11:38' UNTIL '2022-01-06 12:00' WHERE id=1\"",
	Args:    cobra.MinimumNArgs(1),
	RunE:    auditSQL,
}

func init() {
	auditCmd.AddCommand(auditSQLCmd)
}

func auditSQL(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	jr, err := immudb.NewJsonSQLRepository(immuCli, args[0])
	if err != nil {
		return fmt.Errorf("could not create json sql repository, %w", err)
	}

	query := ""
	if len(args) == 2 {
		query = args[1]
	}

	history, err := jr.History(query)
	if err != nil {
		return fmt.Errorf("could not get audit, %w", err)
	}

	for _, h := range history {
		fmt.Println(string(h))
	}
	return nil
}
