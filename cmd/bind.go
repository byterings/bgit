package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/byterings/bgit/core/config"
	coreidentity "github.com/byterings/bgit/core/identity"
	corerepo "github.com/byterings/bgit/core/repo"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	bindUser   string
	bindForce  bool
	bindRemove bool
)

var bindCmd = &cobra.Command{
	Use:   "bind",
	Short: "Bind current repository to an identity",
	Long: `Bind the current repository to a specific identity.

The binding persists regardless of the global active user. When you work in a bound
repository, bgit commands will use the bound identity.

Examples:
  bgit bind                  # Bind to current active user
  bgit bind --user work      # Bind to specific user
  bgit bind --force          # Override existing binding
  bgit bind --remove         # Remove binding`,
	RunE: runBind,
}

func init() {
	rootCmd.AddCommand(bindCmd)
	bindCmd.Flags().StringVarP(&bindUser, "user", "u", "", "User alias to bind to (default: active user)")
	bindCmd.Flags().BoolVarP(&bindForce, "force", "f", false, "Override existing binding")
	bindCmd.Flags().BoolVarP(&bindRemove, "remove", "r", false, "Remove binding for current repository")
}

func runBind(cmd *cobra.Command, args []string) error {
	if err := autoInit(); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	repoRoot := coreidentity.FindGitRoot(cwd)
	if repoRoot == "" {
		return fmt.Errorf("not in a git repository. Run this command from inside a git repo.")
	}

	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if bindRemove {
		return removeBind(cfg, repoRoot)
	}

	userAlias := bindUser
	if userAlias == "" {
		userAlias = cfg.ActiveUser
	}

	if userAlias == "" {
		return fmt.Errorf("no active user set. Use --user flag or run 'bgit use <alias>' first")
	}

	result, err := corerepo.BindRepository(cfg, repoRoot, userAlias, bindForce)
	if err != nil {
		return fmt.Errorf("failed to add binding: %w", err)
	}

	if result.NoChange {
		ui.Info(fmt.Sprintf("Repository already bound to '%s'. No changes needed.", userAlias))
		return nil
	}

	if result.ExistingBinding != nil && result.ExistingBinding.User != userAlias {
		if !bindForce {
			return fmt.Errorf("repository already bound to '%s'. Use --force to override", result.ExistingBinding.User)
		}
		ui.Warning(fmt.Sprintf("Overriding existing binding from '%s' to '%s'", result.ExistingBinding.User, userAlias))
	}
	if result.Workspace != nil {
		ui.Warning(fmt.Sprintf("Note: This repo is inside workspace '%s' which uses '%s'", result.Workspace.Path, result.Workspace.User))
		ui.Info("Explicit binding takes precedence over workspace.")
	}
	ui.Success(fmt.Sprintf("Bound repository to '%s' (%s)", userAlias, result.User.GitHubUsername))
	fmt.Printf("  Path: %s\n", repoRoot)
	fmt.Printf("  Email: %s\n", result.User.Email)

	return nil
}

func removeBind(cfg *config.Config, repoRoot string) error {
	result, err := corerepo.RemoveBinding(cfg, repoRoot)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	if result.Binding == nil {
		ui.Info("No binding found for this repository. Using global active user.")
		return nil
	}

	previousUser := result.Binding.User

	if result.Changed {
		ui.Success(fmt.Sprintf("Removed binding for '%s'", previousUser))
		ui.Info("Repository will now use workspace identity (if inside one) or global active user.")
	}

	return nil
}
