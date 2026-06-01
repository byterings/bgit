package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/byterings/bgit/core/config"
	coreidentity "github.com/byterings/bgit/core/identity"
	corerepo "github.com/byterings/bgit/core/repo"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var cloneNoBind bool

var cloneCmd = &cobra.Command{
	Use:   "clone <url> [directory]",
	Short: "Clone a repository with the correct SSH configuration",
	Long: `Clone a GitHub repository using the active user's SSH configuration.

Accepts any GitHub URL format (HTTPS or SSH) and automatically converts it
to use the correct SSH host alias for the active user.`,
	Example: `  # Clone using HTTPS URL
  bgit clone https://github.com/user/repo.git

  # Clone using SSH URL
  bgit clone git@github.com:user/repo.git

  # Clone to specific directory
  bgit clone https://github.com/user/repo.git my-folder`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().BoolVar(&cloneNoBind, "no-bind", false, "Skip automatic repository binding after clone")
}

func runClone(cmd *cobra.Command, args []string) error {
	url := args[0]
	var directory string
	if len(args) > 1 {
		directory = args[1]
	}

	// Auto-initialize if needed
	if err := autoInit(); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve effective identity (workspace > binding > global)
	resolution, err := coreidentity.GetEffectiveResolution(cfg)
	if err != nil || resolution == nil || resolution.User == nil {
		// Fall back to checking global active user
		if cfg.ActiveUser == "" {
			return fmt.Errorf("no active user set\nRun: bgit use <alias>")
		}
		resolution = &coreidentity.Resolution{
			User:   cfg.FindUserByAlias(cfg.ActiveUser),
			Alias:  cfg.ActiveUser,
			Source: coreidentity.SourceGlobal,
		}
		if resolution.User == nil {
			return fmt.Errorf("active user '%s' not found in config", cfg.ActiveUser)
		}
	}

	activeUser := resolution.User

	// Show identity source if not global
	if resolution.Source != coreidentity.SourceGlobal {
		sourceInfo := ""
		switch resolution.Source {
		case coreidentity.SourceWorkspace:
			sourceInfo = fmt.Sprintf(" (workspace: %s)", resolution.Path)
		case coreidentity.SourceBinding:
			sourceInfo = " (bound repo)"
		}
		ui.Info(fmt.Sprintf("Using identity from %s%s", resolution.Source, sourceInfo))
	}

	// Check if SSH key is configured
	if activeUser.SSHKeyPath == "" {
		ui.Warning("No SSH key configured for this user")
		fmt.Println("Clone may fail. Run: bgit update " + activeUser.Alias + " --ssh-key <path>")
		fmt.Println()
	} else {
		// Ensure SSH agent has the key loaded
		coressh.EnsureKeyLoaded(activeUser)
	}

	// Convert URL to bgit format (uses GitHub username for SSH host)
	convertedURL, err := corerepo.ConvertToBgitURL(url, activeUser.GitHubUsername)
	if err != nil {
		return err
	}

	fmt.Printf("Cloning as: %s\n", activeUser.Alias)
	fmt.Printf("URL: %s\n\n", convertedURL)

	// Build git clone command
	gitArgs := []string{"clone", convertedURL}
	if directory != "" {
		gitArgs = append(gitArgs, directory)
	}

	// Execute git clone
	gitCmd := exec.Command("git", gitArgs...)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	gitCmd.Stdin = os.Stdin

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Println()
	ui.Success("Repository cloned successfully!")

	if !cloneNoBind {
		if _, err := corerepo.BindClonedRepository(cfg, url, directory, resolution.Alias); err != nil {
			ui.Warning(fmt.Sprintf("Clone succeeded, but auto-bind failed: %v", err))
			ui.Info("You can bind manually with: bgit bind --user " + resolution.Alias)
		} else {
			ui.Success(fmt.Sprintf("Repository bound to '%s'", resolution.Alias))
		}
	}

	return nil
}
