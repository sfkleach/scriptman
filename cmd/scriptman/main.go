package main

import (
	"fmt"
	"os"

	"github.com/sfkleach/scriptman/pkg/install"
	"github.com/sfkleach/scriptman/pkg/version"
	"github.com/spf13/cobra"
)

var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "scriptman",
	Short: "Scriptman - Script manager",
	Long:  `Scriptman is a command-line tool for managing scripts from GitHub repositories.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			if err := version.ShowVersion(false, false); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&versionFlag, "version", false, "Print version information")

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(install.NewInstallCommand())
	rootCmd.AddCommand(newInfoCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newCheckCommand())
	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newRemoveCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// newInfoCommand creates the info command stub.
func newInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show information about an installed script",
		Long:  "Show information about an installed script (TBD).",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("info command: TBD")
		},
	}
}

// newListCommand creates the list command stub.
func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed scripts",
		Long:  "List all installed scripts (TBD).",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("list command: TBD")
		},
	}
}

// newCheckCommand creates the check command stub.
func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check for updates to installed scripts",
		Long:  "Check for updates to installed scripts (TBD).",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("check command: TBD")
		},
	}
}

// newUpdateCommand creates the update command stub.
func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update installed scripts",
		Long:  "Update installed scripts to the latest versions (TBD).",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("update command: TBD")
		},
	}
}

// newRemoveCommand creates the remove command stub.
func newRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove an installed script",
		Long:  "Remove an installed script (TBD).",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("remove command: TBD")
		},
	}
}
