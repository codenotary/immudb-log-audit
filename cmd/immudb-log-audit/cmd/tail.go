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
