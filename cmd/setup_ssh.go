package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/byterings/bgit/core/config"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var setupSSHCmd = &cobra.Command{
	Use:   "setup-ssh",
	Short: "Setup SSH agent (Windows helper)",
	Long: `Setup SSH agent and add SSH keys.
This is especially useful on Windows where SSH agent needs to be started manually.`,
	Deprecated: "use 'bgit setup' instead",
	RunE:       runSetupSSH,
}

func init() {
	rootCmd.AddCommand(setupSSHCmd)
}

func runSetupSSH(cmd *cobra.Command, args []string) error {
	ui.Warning("'bgit setup-ssh' is deprecated. Use 'bgit setup' for full setup.")

	// Auto-initialize if needed
	if err := autoInit(); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Users) == 0 {
		ui.Info("No users configured. Run: bgit add")
		return nil
	}

	fmt.Println("Setting up SSH agent...")
	fmt.Println()

	// Windows-specific setup
	if runtime.GOOS == "windows" {
		if err := setupWindowsSSH(cfg); err != nil {
			return err
		}
	} else {
		if err := setupUnixSSH(cfg); err != nil {
			return err
		}
	}

	fmt.Println()
	ui.Success("SSH setup complete!")
	fmt.Println()
	fmt.Println("Test your connection:")
	fmt.Println("  ssh -T git@github.com")
	fmt.Println()

	return nil
}

func setupWindowsSSH(cfg *config.Config) error {
	fmt.Println("Windows SSH Setup:")
	fmt.Println()

	fmt.Println("1. Starting ssh-agent service...")
	coressh.StartSSHAgent()
	ui.Success("ssh-agent service start attempted")

	// Add keys to ssh-agent
	fmt.Println()
	fmt.Println("2. Adding SSH keys to agent...")

	report := coressh.SetupAgentForUsers(cfg.Users)
	addedCount := 0
	for alias, keyPath := range report.Added {
		fmt.Printf("   Added key: %s\n", keyPath)
		ui.Success(fmt.Sprintf("Added key for %s", alias))
		addedCount++
	}
	for alias, output := range report.Failed {
		ui.Error(fmt.Sprintf("Failed to add key for %s", alias))
		if output != "" {
			fmt.Printf("   Error: %s\n", strings.TrimSpace(output))
		}
	}

	fmt.Println()
	fmt.Printf("Added %d SSH keys to agent\n", addedCount)

	// List loaded keys
	fmt.Println()
	fmt.Println("3. Verifying loaded keys...")
	output, err := coressh.ListAgentKeys()
	if err != nil {
		ui.Info("No keys currently loaded in ssh-agent")
	} else {
		fmt.Println(output)
	}

	return nil
}

func setupUnixSSH(cfg *config.Config) error {
	fmt.Println("Unix/Linux SSH Setup:")
	fmt.Println()

	// Check if ssh-agent is running
	if !coressh.IsSSHAgentRunning() {
		fmt.Println("1. Starting ssh-agent...")
		fmt.Println("   Run: eval $(ssh-agent)")
		fmt.Println()
	} else {
		ui.Success("ssh-agent is running")
		fmt.Println()
	}

	// Add keys
	fmt.Println("2. Adding SSH keys to agent...")

	report := coressh.SetupAgentForUsers(cfg.Users)
	addedCount := 0
	for alias, keyPath := range report.Added {
		fmt.Printf("   Added key: %s\n", keyPath)
		ui.Success(fmt.Sprintf("Added key for %s", alias))
		addedCount++
	}
	for alias, output := range report.Failed {
		ui.Error(fmt.Sprintf("Failed to add key for %s", alias))
		if output != "" {
			fmt.Printf("   Error: %s\n", strings.TrimSpace(output))
		}
	}

	fmt.Println()
	fmt.Printf("Added %d SSH keys to agent\n", addedCount)

	// List loaded keys
	fmt.Println()
	fmt.Println("3. Verifying loaded keys...")
	output, err := coressh.ListAgentKeys()
	if err != nil {
		ui.Info("No keys currently loaded in ssh-agent")
	} else {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Println("  ", line)
			}
		}
	}

	return nil
}
