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
	"strings"

	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	"github.com/spf13/cobra"
)

var readKVCmd = &cobra.Command{
	Use:   "kv <collection> <<indexed field=value prefix>>",
	Short: "Read audit data from immudb key-value collection.",
	Example: `immudb-log-audit read kv samplecollection
immudb-log-audit read kv samplecollection indexed_field1=prefix1
immudb-log-audit read kv samplecollection indexed_field2=prefix2`,
	RunE: readKV,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	readCmd.AddCommand(readKVCmd)
}

func readKV(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	jr, err := immudb.NewJsonKVRepository(immuCli, args[0])
	if err != nil {
		return fmt.Errorf("could not create json kv repository, %w", err)
	}

	key := ""
	prefix := ""
	if len(args) == 2 {
		split := strings.SplitN(args[1], "=", 2)
		key = split[0]

		if len(split) > 1 {
			prefix = split[1]
		}
	}

	jsons, err := jr.Read(key, prefix)
	if err != nil {
		return fmt.Errorf("could not read, %w", err)
	}

	for _, j := range jsons {
		fmt.Println(string(j))
	}

	return nil
}
