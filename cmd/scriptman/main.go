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
	// TODO: Add install, list, remove, check, update commands.
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
