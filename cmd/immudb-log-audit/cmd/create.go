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
	"github.com/spf13/cobra"
)

var flagParser string
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create collection in immudb",
	RunE:  create,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().StringVar(&flagParser, "parser", "", "Line parser to be used. When not specified, lines will be considered as jsons. Also available 'pgaudit', 'wrap'. For those, indexes are predefined.")
}

func create(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "create" {
		return cmd.Help()
	}

	return runParentCmdE(cmd, args)
}
