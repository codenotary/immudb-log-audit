package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/codenotary/immudb/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var immuCli client.ImmuClient

var rootCmd = &cobra.Command{
	Use:               "immudb-log-audit",
	Short:             "Store and audit your data in immudb",
	RunE:              root,
	PersistentPostRun: rootPost,
}

var usageTemplate = `Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableFlags}}

Flags:
{{.Flags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func init() {
	rootCmd.SetUsageTemplate(usageTemplate)
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
