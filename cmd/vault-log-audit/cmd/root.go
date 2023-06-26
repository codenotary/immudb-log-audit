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
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	vaultclient "github.com/codenotary/immudb-log-audit/pkg/client/vault"
	"github.com/codenotary/immudb-log-audit/pkg/cmd"
	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version   string
	Commit    string
	BuildTime string
)

var vaultClient vaultclient.ClientWithResponsesInterface
var ledger string
var flagParser string
var flagBatchMode bool

func version() string {
	return fmt.Sprintf("%s, commit: %s, build time: %s",
		Version, Commit,
		time.Unix(func() int64 {
			i, _ := strconv.ParseInt(BuildTime, 10, 64)
			return i
		}(), 0))
}

var rootCmd = &cobra.Command{
	Use:     "vault-log-audit",
	Short:   "Store and audit your data in immudb vault",
	RunE:    root,
	Version: version(),
}

func init() {
	cobra.OnInitialize(func() {
		replacer := strings.NewReplacer("-", "_")
		viper.SetEnvKeyReplacer(replacer)
		viper.AutomaticEnv()
	})

	rootCmd.SetUsageTemplate(cmd.UsageTemplate)
	rootCmd.PersistentFlags().String("vault-address", "https://vault.immudb.io/", "vault address, can be set with VAULT_ADDRESS env var")
	rootCmd.PersistentFlags().String("vault-api-key", "", "Vault api key, can be set with VAULT_API_KEY env var")
	rootCmd.PersistentFlags().StringVar(&ledger, "ledger", "default", "Ledger to be used")
	rootCmd.PersistentFlags().StringVar(&flagParser, "parser", "", "Line parser to be used. When not specified, lines will be considered as jsons. Also available 'pgaudit', 'pgauditjsonlog', 'wrap'. For those, indexes are predefined.")
	rootCmd.PersistentFlags().BoolVar(&flagBatchMode, "batch-mode", true, "")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (trace, debug, info, warn, error)")
}

func root(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "vault-log-audit" {
		return cmd.Help()
	}

	logLevelString, _ := cmd.Flags().GetString("log-level")
	logLevel, err := log.ParseLevel(logLevelString)
	if err != nil {
		return err
	}

	log.SetLevel(logLevel)

	vaultAddress := viper.GetString("vault-address")
	if vaultAddress == "" {
		vaultAddress = "https://vault.immudb.io/"
	}

	vaultAPIKey := viper.GetString("vault-api-key")

	if vaultAPIKey == "" {
		return errors.New("vault-api-key cannot be empty")
	}

	apikeyProvider, err := securityprovider.NewSecurityProviderApiKey("header", "X-API-Key", vaultAPIKey)
	if err != nil {
		return fmt.Errorf("could not configure API Key provider, %w", err)
	}

	vaultAddress, _ = url.JoinPath(vaultAddress, "ics/api/v1")

	vaultClient, err = vaultclient.NewClientWithResponses(vaultAddress, vaultclient.WithRequestEditorFn(apikeyProvider.Intercept))
	if err != nil {
		return fmt.Errorf("could not initialize vault client, %w", err)
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runParentCmdE(cmd *cobra.Command, args []string) error {
	if cmd.Parent() != nil && cmd.Parent().RunE != nil {
		err := cmd.Parent().RunE(cmd.Parent(), args)
		if err != nil {
			return err
		}
	}

	return nil
}
