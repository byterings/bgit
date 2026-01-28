package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/identity"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current identity status",
	Long: `Display the current identity status including:
- Active global identity
- Current repository binding (if in a git repo)
- Effective identity for current location
- Configured workspaces and bindings

This helps you understand which identity will be used for git operations.`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	removed := cfg.CleanupInvalidPaths()
	if removed > 0 {
		if err := config.SaveConfig(cfg); err != nil {
			ui.Warning(fmt.Sprintf("Failed to save config after cleanup: %v", err))
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}

	var resolution *identity.Resolution
	if cwd != "" {
		resolution, _ = identity.ResolveIdentity(cfg, cwd)
	}

	printActiveIdentity(cfg, resolution)
	printCurrentRepo(cfg, cwd, resolution)
	printWorkspaces(cfg)
	printBindings(cfg)

	return nil
}

func printActiveIdentity(cfg *config.Config, resolution *identity.Resolution) {
	fmt.Println()
	fmt.Println("Active Identity")
	fmt.Println("───────────────")

	if cfg.ActiveUser == "" {
		fmt.Println("  No active user set")
		fmt.Println("  Run 'bgit use <alias>' to set one")
		return
	}

	user := cfg.FindUserByAlias(cfg.ActiveUser)
	if user == nil {
		fmt.Printf("  Active user '%s' not found in config\n", cfg.ActiveUser)
		return
	}

	fmt.Printf("  Name:     %s\n", user.Name)
	fmt.Printf("  Email:    %s\n", user.Email)
	fmt.Printf("  GitHub:   %s\n", user.GitHubUsername)

	// Check SSH key status
	sshStatus := "✓"
	if user.SSHKeyPath != "" {
		if _, err := os.Stat(user.SSHKeyPath); os.IsNotExist(err) {
			sshStatus = "✗ (missing)"
		}
	} else {
		sshStatus = "⚠ (not configured)"
	}
	fmt.Printf("  SSH Key:  %s %s\n", user.SSHKeyPath, sshStatus)
}

func printCurrentRepo(cfg *config.Config, cwd string, resolution *identity.Resolution) {
	fmt.Println()
	fmt.Println("Current Location")
	fmt.Println("────────────────")

	if cwd == "" {
		fmt.Println("  Could not determine current directory")
		return
	}

	repoRoot := identity.FindGitRoot(cwd)

	if repoRoot == "" {
		fmt.Printf("  Path: %s\n", cwd)
		fmt.Println("  Not inside a git repository")
	} else {
		fmt.Printf("  Path: %s\n", repoRoot)
	}

	if resolution != nil {
		fmt.Println()
		fmt.Println("Effective Identity")
		fmt.Println("──────────────────")

		sourceStr := ""
		switch resolution.Source {
		case identity.SourceWorkspace:
			sourceStr = fmt.Sprintf("(workspace: %s)", resolution.Path)
		case identity.SourceBinding:
			sourceStr = fmt.Sprintf("(bound repo)")
		case identity.SourceGlobal:
			sourceStr = "(global)"
		}

		fmt.Printf("  Using: %s %s\n", resolution.Alias, sourceStr)

		if resolution.User != nil {
			fmt.Printf("  Email: %s\n", resolution.User.Email)
			fmt.Printf("  GitHub: %s\n", resolution.User.GitHubUsername)
		}

		if cfg.ActiveUser != "" && resolution.Alias != cfg.ActiveUser && resolution.Source != identity.SourceGlobal {
			fmt.Println()
			ui.Warning("Identity mismatch!")
			fmt.Printf("  Global active: %s\n", cfg.ActiveUser)
			fmt.Printf("  Effective:     %s\n", resolution.Alias)
			ui.Info("The effective identity will be used for bgit commands in this location.")
		}
	}
}

func printWorkspaces(cfg *config.Config) {
	workspaces := cfg.GetWorkspaces()
	if len(workspaces) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("Workspaces")
	fmt.Println("──────────")

	for _, ws := range workspaces {
		status := "✓"
		if _, err := os.Stat(ws.Path); os.IsNotExist(err) {
			status = "✗"
		}
		fmt.Printf("  %s %s → %s\n", status, shortenPath(ws.Path), ws.User)
	}
}

func printBindings(cfg *config.Config) {
	bindings := cfg.GetBindings()
	if len(bindings) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("Bound Repositories")
	fmt.Println("──────────────────")

	for _, b := range bindings {
		status := "✓"
		if _, err := os.Stat(b.Path); os.IsNotExist(err) {
			status = "✗"
		}
		fmt.Printf("  %s %s → %s\n", status, shortenPath(b.Path), b.User)
	}
}

// shortenPath shortens home directory paths with ~
func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	if len(absPath) > len(home) && absPath[:len(home)] == home {
		return "~" + absPath[len(home):]
	}

	return path
}
