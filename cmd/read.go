/*
Copyright 2022 Codenotary Inc. All rights reserved.

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

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read audit data from immudb.",
	RunE:  read,
}

func init() {
	rootCmd.AddCommand(readCmd)
}

func read(cmd *cobra.Command, args []string) error {
	if cmd.CalledAs() == "read" {
		return cmd.Help()
	}

	err := runParentCmdE(cmd, args)
	if err != nil {
		return err
	}

	return nil
}
