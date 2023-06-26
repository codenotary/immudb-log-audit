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

var readSQLCmd = &cobra.Command{
	Use:   "sql <collection> <<SQL query conditions>>",
	Short: "Read audit data from immudb SQL collection.",
	RunE:  readSQL,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	readCmd.AddCommand(readSQLCmd)
}

func readSQL(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	jr, err := immudb.NewJsonSQLRepository(immuCli, args[0])
	if err != nil {
		return fmt.Errorf("could not create json kv repository, %w", err)
	}

	query := ""
	if len(args) == 2 {
		query = args[1]
	}

	jsons, err := jr.Read(query)
	if err != nil {
		return fmt.Errorf("could not read, %w", err)
	}

	for _, j := range jsons {
		fmt.Println(string(j))
	}

	return nil
}
