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
	"encoding/json"
	"fmt"

	vaultclient "github.com/codenotary/immudb-log-audit/pkg/client/vault"
	"github.com/codenotary/immudb-log-audit/pkg/repository/vault"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <collection>",
	Short: "create collection in immudb vault",
	RunE:  create,
	Args:  cobra.MaximumNArgs(1),
}

var indexesFlag string

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&indexesFlag, "indexes", "", "immudb vault indexes configuration in JSON format")
}

func create(cmd *cobra.Command, args []string) error {
	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	collection := "default"
	if len(args) == 1 {
		collection = args[0]
	} else {
		log.Info("Creating default collection")
	}

	var createRequest *vaultclient.CollectionCreateRequest
	if indexesFlag != "" {
		err := json.Unmarshal([]byte(indexesFlag), &createRequest)
		if err != nil {
			return fmt.Errorf("invalid indexes configuration, %w", err)
		}
	} else if flagParser == "pgaudit" {
		statementIDType := vaultclient.INTEGER
		subStatementIDType := vaultclient.INTEGER
		timestampType := vaultclient.STRING
		auditTypeType := vaultclient.STRING
		classType := vaultclient.STRING
		commandType := vaultclient.STRING
		createRequest = &vaultclient.CollectionCreateRequest{
			Fields: &[]vaultclient.Field{
				{
					Name: "statement_id",
					Type: &statementIDType,
				},
				{
					Name: "substatement_id",
					Type: &subStatementIDType,
				},
				{
					Name: "timestamp",
					Type: &timestampType,
				},
				{
					Name: "audit_type",
					Type: &auditTypeType,
				},
				{
					Name: "class",
					Type: &classType,
				},
				{
					Name: "command",
					Type: &commandType,
				},
			},
			Indexes: &[]vaultclient.Index{
				{
					Fields: []string{"statement_id"},
				},
				{
					Fields: []string{"substatement_id"},
				},
				{
					Fields: []string{"audit_type"},
				},
				{
					Fields: []string{"class"},
				},
				{
					Fields: []string{"command"},
				},
			},
		}
	} else if flagParser == "pgauditjsonlog" {
		dbNameType := vaultclient.STRING
		userType := vaultclient.STRING
		statementIDType := vaultclient.INTEGER
		subStatementIDType := vaultclient.INTEGER
		timestampType := vaultclient.STRING
		auditTypeType := vaultclient.STRING
		classType := vaultclient.STRING
		commandType := vaultclient.STRING
		createRequest = &vaultclient.CollectionCreateRequest{
			Fields: &[]vaultclient.Field{
				{
					Name: "user",
					Type: &userType,
				},
				{
					Name: "dbname",
					Type: &dbNameType,
				},
				{
					Name: "statement_id",
					Type: &statementIDType,
				},
				{
					Name: "substatement_id",
					Type: &subStatementIDType,
				},
				{
					Name: "timestamp",
					Type: &timestampType,
				},
				{
					Name: "audit_type",
					Type: &auditTypeType,
				},
				{
					Name: "class",
					Type: &classType,
				},
				{
					Name: "command",
					Type: &commandType,
				},
			},
			Indexes: &[]vaultclient.Index{
				{
					Fields: []string{"dbname"},
				},
				{
					Fields: []string{"statement_id"},
				},
				{
					Fields: []string{"substatement_id"},
				},
				{
					Fields: []string{"audit_type"},
				},
				{
					Fields: []string{"class"},
				},
				{
					Fields: []string{"command"},
				},
			},
		}
	} else if flagParser == "wrap" {
		uidType := vaultclient.STRING
		timestampType := vaultclient.STRING
		createRequest = &vaultclient.CollectionCreateRequest{
			Fields: &[]vaultclient.Field{
				{
					Name: "uid",
					Type: &uidType,
				},
				{
					Name: "log_timestamp",
					Type: &timestampType,
				},
			},
			Indexes: &[]vaultclient.Index{
				{
					Fields: []string{"uid"},
				},
				{
					Fields: []string{"log_timestamp"},
				},
			},
		}
	}

	err = vault.SetupJsonObjectRepository(vaultClient, ledger, collection, createRequest)
	if err != nil {
		return fmt.Errorf("could not set up collection, %w", err)
	}

	return nil
}
