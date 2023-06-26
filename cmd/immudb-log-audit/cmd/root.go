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
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/codenotary/immudb-log-audit/pkg/cmd"
	"github.com/codenotary/immudb/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Version   string
	Commit    string
	BuildTime string
)

var immuCli client.ImmuClient

func version() string {
	return fmt.Sprintf("%s, commit: %s, build time: %s",
		Version, Commit,
		time.Unix(func() int64 {
			i, _ := strconv.ParseInt(BuildTime, 10, 64)
			return i
		}(), 0))
}

var rootCmd = &cobra.Command{
	Use:               "immudb-log-audit",
	Short:             "Store and audit your data in immudb",
	RunE:              root,
	PersistentPostRun: rootPost,
	Version:           version(),
}

func init() {
	rootCmd.SetUsageTemplate(cmd.UsageTemplate)
	rootCmd.PersistentFlags().String("immudb-host", "localhost", "immudb host")
	rootCmd.PersistentFlags().Int("immudb-port", 3322, "immudb port")
	rootCmd.PersistentFlags().String("immudb-database", "defaultdb", "immudb database")
	rootCmd.PersistentFlags().String("immudb-user", "immudb", "immudb user")
	rootCmd.PersistentFlags().String("immudb-password", "immudb", "immudb user password")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (trace, debug, info, warn, error)")
}

func root(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "immudb-log-audit" {
		return cmd.Help()
	}

	logLevelString, _ := cmd.Flags().GetString("log-level")
	logLevel, err := log.ParseLevel(logLevelString)
	if err != nil {
		return err
	}

	log.SetLevel(logLevel)

	immudbHost, _ := cmd.Flags().GetString("immudb-host")
	immudbPort, _ := cmd.Flags().GetInt("immudb-port")
	immudbDb, _ := cmd.Flags().GetString("immudb-database")
	immudbUser, _ := cmd.Flags().GetString("immudb-user")
	immudbPassword, _ := cmd.Flags().GetString("immudb-password")

	opts := client.DefaultOptions().WithAddress(immudbHost).WithPort(immudbPort)
	immuCli = client.NewClient().WithOptions(opts)

	err = immuCli.OpenSession(context.TODO(), []byte(immudbUser), []byte(immudbPassword), immudbDb)
	if err != nil {
		return err
	}

	return nil
}

func rootPost(cmd *cobra.Command, args []string) {
	if immuCli != nil {
		immuCli.CloseSession(context.TODO())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
