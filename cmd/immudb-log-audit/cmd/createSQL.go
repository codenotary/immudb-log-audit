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
	"strings"

	"github.com/codenotary/immudb-log-audit/pkg/repository/immudb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createSQLCmd = &cobra.Command{
	Use:   "sql <collection>",
	Short: "Create collection in immudb with SQL",
	RunE:  createSQL,
	Args:  cobra.ExactArgs(1),
}

func init() {
	createCmd.AddCommand(createSQLCmd)
	createSQLCmd.Flags().StringSlice("primary-key", nil, "List of columns to be used as primary key")
	createSQLCmd.Flags().StringSlice("columns", nil, "List of fields to be used as columns")
}

func createSQL(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	primaryKey, _ := cmd.Flags().GetStringSlice("primary-key")
	flagColumns, _ := cmd.Flags().GetStringSlice("columns")
	if flagParser == "pgaudit" {
		flagColumns = []string{"id=INTEGER AUTO_INCREMENT", "statement_id=INTEGER", "substatement_id=INTEGER", "server_timestamp=TIMESTAMP", "timestamp=TIMESTAMP", "audit_type=VARCHAR[256]", "class=VARCHAR[256]", "command=VARCHAR[256]"}
		primaryKey = []string{"id"}
		log.WithField("columns", flagColumns).WithField("primary_key", primaryKey).Info("Using default indexes for pgaudit parser")
	} else if flagParser == "pgauditjsonlog" {
		flagColumns = []string{"id=INTEGER AUTO_INCREMENT", "user=VARCHAR[256]", "dbname=VARCHAR[256]", "session_id=VARCHAR[256]", "statement_id=INTEGER", "substatement_id=INTEGER", "server_timestamp=TIMESTAMP", "timestamp=TIMESTAMP", "audit_type=VARCHAR[256]", "class=VARCHAR[256]", "command=VARCHAR[256]"}
		primaryKey = []string{"id"}
		log.WithField("columns", flagColumns).WithField("primary_key", primaryKey).Info("Using default indexes for pgauditjsonlog parser")
	} else if flagParser == "wrap" {
		flagColumns = []string{"uid=VARCHAR[36]", "log_timestamp=TIMESTAMP"}
		primaryKey = []string{"uid"}
		log.WithField("columns", flagColumns).WithField("primary_key", primaryKey).Info("Using default indexes for wrap parser")
	} else if flagParser != "" {
		return fmt.Errorf("unkown parser %s", flagParser)
	}

	if len(flagColumns) == 0 || len(primaryKey) == 0 {
		return errors.New("at least one column and primary key needs to be specified")
	}

	err = immudb.NewConfigs(immuCli).WriteTypeParser("sql", args[0], flagParser)
	if err != nil {
		return fmt.Errorf("could not create json repository parser config, %w", err)
	}
	immudb.SetupJsonSQLRepository(immuCli, args[0], strings.Join(primaryKey, ","), flagColumns)

	return nil
}
