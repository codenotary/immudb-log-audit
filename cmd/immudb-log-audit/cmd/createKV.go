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
	"errors"
	"fmt"

	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createKVCmd = &cobra.Command{
	Use:   "kv <collection>",
	Short: "Create collection in immudb with key-value",
	Example: `immudb-log-audit create kv samplecollection --parser pgaudit
immudb-log-audit create kv samplecollection --indexes unique_field1,field2,field3
immudb-log-audit create kv samplecollection --indexes field1+field2,field2,field3`,
	RunE: createKV,
	Args: cobra.ExactArgs(1),
}

func init() {
	createCmd.AddCommand(createKVCmd)
	createKVCmd.Flags().StringSlice("indexes", nil, "List of JSON fields to create indexes for. First entry is considered as unique primary key. If needed, multiple fields can be used as primary key with syntax field1+field2...")
}

func createKV(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	flagIndexes, _ := cmd.Flags().GetStringSlice("indexes")
	if flagParser == "pgaudit" {
		flagIndexes = []string{"uid", "statement_id", "substatement_id", "server_timestamp", "timestamp", "audit_type", "class", "command"}
		log.WithField("indexes", flagIndexes).Info("Using default indexes for pgaudit parser")
	} else if flagParser == "pgauditjsonlog" {
		flagIndexes = []string{"uid", "user", "dbname", "session_id", "statement_id", "substatement_id", "server_timestamp", "timestamp", "audit_type", "class", "command"}
		log.WithField("indexes", flagIndexes).Info("Using default indexes for pgauditjsonlog parser")
	} else if flagParser == "wrap" {
		flagIndexes = []string{"uid", "timestamp"}
		log.WithField("indexes", flagIndexes).Info("Using default indexes for wrap parser")
	} else if flagParser != "" {
		return fmt.Errorf("unkown parser %s", flagParser)
	}

	if len(flagIndexes) == 0 {
		return errors.New("at least primary key needs to be specified")
	}

	err = immudb.NewConfigs(immuCli).WriteTypeParser("kv", args[0], flagParser)
	if err != nil {
		return fmt.Errorf("could not create json repository parser config, %w", err)
	}

	err = immudb.SetupJsonKVRepository(immuCli, args[0], flagIndexes)
	if err != nil {
		return fmt.Errorf("could not create json repository, %w", err)
	}

	return nil
}
