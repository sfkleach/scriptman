package list

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/sfkleach/scriptman/pkg/config"
	"github.com/sfkleach/scriptman/pkg/registry"
	"github.com/spf13/cobra"
)

// Options contains the list command options.
type Options struct {
	Long       bool
	JSONOutput bool
}

// NewListCommand creates the list command.
func NewListCommand() *cobra.Command {
	var opts Options

	cmd := &cobra.Command{
		Use:     "list [executable-name]",
		Aliases: []string{"ls"},
		Short:   "List installed scripts",
		Long:    "Display executables managed by scriptman. Optionally filters by name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(args, &opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Long, "long", "l", false, "Show detailed information")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Output as JSON")

	return cmd
}

// runList executes the list command.
func runList(args []string, opts *Options) error {
	// Load registry.
	registryPath, err := config.GetDefaultRegistryPath()
	if err != nil {
		return fmt.Errorf("failed to get registry path: %w", err)
	}

	reg, err := registry.Load(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Filter by name if provided.
	var filterName string
	if len(args) > 0 {
		filterName = args[0]
	}

	// Build filtered list.
	filtered := make(map[string]*registry.Script)
	for name, script := range reg.Scripts {
		if filterName == "" || name == filterName {
			filtered[name] = script
		}
	}

	// Check if any scripts match.
	if len(filtered) == 0 {
		if filterName != "" {
			fmt.Printf("No scripts matching '%s'.\n", filterName)
		} else {
			fmt.Println("No scripts installed.")
		}
		return nil
	}

	// Handle JSON output.
	if opts.JSONOutput {
		return outputJSON(filtered)
	}

	// Get sorted list of script names.
	names := make([]string, 0, len(filtered))
	for name := range filtered {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print scripts.
	if opts.Long {
		outputLong(names, filtered)
	} else {
		outputShort(names, filtered)
	}

	return nil
}

// outputShort prints just the executable names, one per line.
func outputShort(names []string, scripts map[string]*registry.Script) {
	for _, name := range names {
		fmt.Println(name)
	}
}

// outputLong prints detailed information about scripts.
func outputLong(names []string, scripts map[string]*registry.Script) {
	for i, name := range names {
		if i > 0 {
			fmt.Println()
		}
		script := scripts[name]
		fmt.Printf("Name:         %s\n", name)
		fmt.Printf("Repository:   %s\n", script.Repo)
		fmt.Printf("Source Path:  %s\n", script.SourcePath)
		if script.Version != "" {
			fmt.Printf("Version:      %s\n", script.Version)
		} else {
			fmt.Printf("Version:      (main branch)\n")
		}
		if script.Commit != "" {
			fmt.Printf("Commit:       %s\n", script.Commit[:min(7, len(script.Commit))])
		}
		fmt.Printf("Interpreter:  %s\n", script.Interpreter)
		fmt.Printf("Local Script: %s\n", script.LocalScript)
		fmt.Printf("Wrapper Path: %s\n", script.WrapperPath)
		fmt.Printf("Installed At: %s\n", script.InstalledAt.Format("2006-01-02 15:04:05"))
	}
}

// outputJSON prints scripts as JSON.
func outputJSON(scripts map[string]*registry.Script) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(scripts)
}
