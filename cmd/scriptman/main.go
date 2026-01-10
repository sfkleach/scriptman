package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	rootCmd.AddCommand(newInstallCommand())
	rootCmd.AddCommand(newInfoCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newCheckCommand())
	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newRemoveCommand())
}

func main() {
	// Check if we're being invoked as a wrapped script (runner mode).
	basename := filepath.Base(os.Args[0])
	if basename != "scriptman" {
		runScript(basename)
		return
	}

	// Management mode.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runScript handles runner mode when invoked via a hardlink.
func runScript(name string) {
	// Find our own location.
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptman: cannot determine executable path: %v\n", err)
		os.Exit(1)
	}
	dir := filepath.Dir(self)

	// Look for companion shell script.
	shScript := filepath.Join(dir, name+".sh")
	if _, err := os.Stat(shScript); err == nil {
		// Execute the shell script.
		// TODO: Use syscall.Exec for proper process replacement.
		fmt.Fprintf(os.Stderr, "scriptman: would exec %s\n", shScript)
		os.Exit(0)
	}

	// No script found.
	fmt.Fprintf(os.Stderr, "scriptman: no dispatch found for '%s'\n", name)
	fmt.Fprintf(os.Stderr, "scriptman: looked for %s\n", shScript)
	os.Exit(1)
}

// newInstallCommand creates the install command stub.
func newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install a script from a GitHub repository",
		Long:  "Install a script from a GitHub repository (TBD).",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("install command: TBD")
		},
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
