package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	workspacePath   string
	workspaceUsers  string
	workspaceList   bool
	workspaceRemove string
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Create and manage workspace directories for identities",
	Long: `Create organized workspace directories for each identity.

All repositories cloned within a workspace folder automatically use that identity,
regardless of the global active user.

Examples:
  bgit workspace                    # Create folders for all users in current directory
  bgit workspace --path ~/code      # Create in specific location
  bgit workspace --users work,oss   # Only specific users
  bgit workspace --list             # Show configured workspaces
  bgit workspace --remove work      # Remove workspace binding`,
	RunE: runWorkspace,
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.Flags().StringVarP(&workspacePath, "path", "p", "", "Directory to create workspace folders in (default: current directory)")
	workspaceCmd.Flags().StringVarP(&workspaceUsers, "users", "u", "", "Comma-separated list of user aliases to create folders for (default: all)")
	workspaceCmd.Flags().BoolVarP(&workspaceList, "list", "l", false, "List configured workspaces")
	workspaceCmd.Flags().StringVarP(&workspaceRemove, "remove", "r", "", "Remove workspace binding for the specified user alias")
}

func runWorkspace(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Handle --list
	if workspaceList {
		return listWorkspaces(cfg)
	}

	// Handle --remove
	if workspaceRemove != "" {
		return removeWorkspace(cfg, workspaceRemove)
	}

	// Create workspaces
	return createWorkspaces(cfg)
}

func listWorkspaces(cfg *config.Config) error {
	workspaces := cfg.GetWorkspaces()

	if len(workspaces) == 0 {
		fmt.Println("No workspaces configured.")
		fmt.Println("\nCreate workspaces with: bgit workspace")
		return nil
	}

	fmt.Println("\nConfigured workspaces:")
	fmt.Println()

	for _, ws := range workspaces {
		user := cfg.FindUserByAlias(ws.User)
		userName := ws.User
		if user != nil {
			userName = fmt.Sprintf("%s (%s)", ws.User, user.GitHubUsername)
		}

		// Check if path exists
		status := "✓"
		if _, err := os.Stat(ws.Path); os.IsNotExist(err) {
			status = "✗ (missing)"
		}

		fmt.Printf("  %s %-20s → %s\n", status, userName, ws.Path)
	}

	fmt.Println()
	return nil
}

func removeWorkspace(cfg *config.Config, userAlias string) error {
	// Find workspace for this user
	var found *config.Workspace
	for _, ws := range cfg.GetWorkspaces() {
		if ws.User == userAlias {
			found = &ws
			break
		}
	}

	if found == nil {
		return fmt.Errorf("no workspace found for user '%s'", userAlias)
	}

	if cfg.RemoveWorkspace(userAlias) {
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		ui.Success(fmt.Sprintf("Removed workspace binding for '%s' at %s", userAlias, found.Path))
		ui.Info("Note: The folder was not deleted. Remove it manually if needed.")
	}

	return nil
}

func createWorkspaces(cfg *config.Config) error {
	// Determine base path
	basePath := workspacePath
	if basePath == "" {
		var err error
		basePath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Convert to absolute path
	basePath, err := filepath.Abs(basePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", basePath)
	}

	// Determine which users to create folders for
	var users []config.User
	if workspaceUsers != "" {
		// Parse comma-separated list
		aliases := strings.Split(workspaceUsers, ",")
		for _, alias := range aliases {
			alias = strings.TrimSpace(alias)
			user := cfg.FindUserByAlias(alias)
			if user == nil {
				return fmt.Errorf("user '%s' not found", alias)
			}
			users = append(users, *user)
		}
	} else {
		// All users
		users = cfg.Users
	}

	if len(users) == 0 {
		return fmt.Errorf("no users configured. Add users with: bgit add")
	}

	fmt.Println("Creating workspace directories...")
	fmt.Println()

	created := 0
	for _, user := range users {
		folderPath := filepath.Join(basePath, user.Alias)

		// Create directory if it doesn't exist
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			if err := os.MkdirAll(folderPath, 0755); err != nil {
				ui.Error(fmt.Sprintf("Failed to create %s: %v", folderPath, err))
				continue
			}
			ui.Success(fmt.Sprintf("Created: %s/", user.Alias))
		} else {
			ui.Info(fmt.Sprintf("Exists: %s/", user.Alias))
		}

		// Add workspace binding
		if err := cfg.AddWorkspace(folderPath, user.Alias); err != nil {
			// Already exists is fine
			if !strings.Contains(err.Error(), "already exists") {
				ui.Warning(fmt.Sprintf("Failed to bind %s: %v", user.Alias, err))
				continue
			}
		}
		created++
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("Auto-bound:")
	for _, user := range users {
		folderPath := filepath.Join(basePath, user.Alias)
		fmt.Printf("  %s/**  →  %s (%s)\n", folderPath, user.Alias, user.GitHubUsername)
	}

	fmt.Println()
	ui.Success("Workspace ready! Clone repos into the appropriate folder.")

	return nil
}
