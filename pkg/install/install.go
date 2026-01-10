package install

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfkleach/scriptman/pkg/config"
	"github.com/sfkleach/scriptman/pkg/github"
	"github.com/sfkleach/scriptman/pkg/interpreter"
	"github.com/sfkleach/scriptman/pkg/registry"
	"github.com/sfkleach/scriptman/pkg/wrapper"
	"github.com/spf13/cobra"
)

// Options contains the install command options.
type Options struct {
	Repo        string
	Path        string
	Interpreter string
	Name        string
	Into        string
}

// NewInstallCommand creates the install command.
func NewInstallCommand() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "install REPO PATH",
		Short: "Install a script from a GitHub repository",
		Long: `Install a script from a GitHub repository.

Examples:
  scriptman install owner/repo scripts/myscript.py
  scriptman install owner/repo scripts/tool.rb --name mytool
  scriptman install owner/repo scripts/app.py --interpreter python3.11
  scriptman install owner/repo scripts/util.sh --into ~/bin`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Repo = args[0]
			opts.Path = args[1]
			return runInstall(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Interpreter, "interpreter", "", "Explicit interpreter command")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name for the wrapper (defaults to script filename without extension)")
	cmd.Flags().StringVar(&opts.Into, "into", "", "Target directory for wrapper (defaults to ~/.local/bin)")

	return cmd
}

// runInstall executes the install command.
func runInstall(opts *Options) error {
	// Determine wrapper name.
	name := opts.Name
	if name == "" {
		// Default to script filename without extension.
		base := filepath.Base(opts.Path)
		name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Check for reserved name.
	if name == "scriptman" {
		return fmt.Errorf("'scriptman' is reserved for the management CLI\nChoose a different name with --name")
	}

	// Determine target directory.
	binDir := opts.Into
	if binDir == "" {
		var err error
		binDir, err = config.GetDefaultBinDir()
		if err != nil {
			return fmt.Errorf("failed to get default bin directory: %w", err)
		}
	}

	// Load registry.
	registryPath := config.GetDefaultRegistryPath()
	reg, err := registry.Load(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Check if name already exists.
	if reg.Exists(name) {
		return fmt.Errorf("script '%s' is already installed\nUse 'scriptman remove %s' first or choose a different name with --name", name, name)
	}

	// Fetch script from GitHub.
	fmt.Printf("Fetching %s from %s...\n", opts.Path, opts.Repo)
	scriptContent, err := github.FetchScript(opts.Repo, opts.Path)
	if err != nil {
		return fmt.Errorf("failed to fetch script: %w", err)
	}

	// Detect interpreter.
	fmt.Println("Detecting interpreter...")
	interpPath, err := interpreter.Detect(opts.Path, scriptContent, opts.Interpreter)
	if err != nil {
		return err
	}
	fmt.Printf("Using interpreter: %s\n", interpPath)

	// Determine script storage location.
	scriptDir, err := config.GetDefaultScriptDir()
	if err != nil {
		return fmt.Errorf("failed to get default script directory: %w", err)
	}
	localScriptPath := filepath.Join(scriptDir, filepath.Base(opts.Path))

	// Save script.
	fmt.Printf("Saving script to %s...\n", localScriptPath)
	if err := github.SaveScript(scriptContent, localScriptPath); err != nil {
		return fmt.Errorf("failed to save script: %w", err)
	}

	// Create wrapper.
	fmt.Println("Creating shell script wrapper...")
	wrapperPath := filepath.Join(binDir, name)
	if err := wrapper.CreateWrapper(interpPath, localScriptPath, wrapperPath); err != nil {
		return fmt.Errorf("failed to create wrapper: %w", err)
	}

	// Add to registry.
	reg.Add(name, &registry.Script{
		Repo:        opts.Repo,
		SourcePath:  opts.Path,
		LocalScript: localScriptPath,
		Interpreter: interpPath,
		WrapperPath: wrapperPath,
		InstalledAt: time.Now(),
	})

	// Save registry.
	if err := reg.Save(registryPath); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	fmt.Printf("\n✓ Installed '%s' successfully\n", name)
	fmt.Printf("  Wrapper: %s\n", wrapperPath)
	fmt.Printf("  Script:  %s\n", localScriptPath)

	return nil
}
