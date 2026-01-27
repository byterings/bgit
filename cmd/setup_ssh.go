package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/ui"
)

var setupSSHCmd = &cobra.Command{
	Use:   "setup-ssh",
	Short: "Setup SSH agent (Windows helper)",
	Long: `Setup SSH agent and add SSH keys.
This is especially useful on Windows where SSH agent needs to be started manually.`,
	RunE: runSetupSSH,
}

func init() {
	rootCmd.AddCommand(setupSSHCmd)
}

func runSetupSSH(cmd *cobra.Command, args []string) error {
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

	// Check if ssh-agent service is running
	fmt.Println("1. Starting ssh-agent service...")

	// Start ssh-agent service
	startCmd := exec.Command("powershell", "-Command", "Start-Service ssh-agent")
	if err := startCmd.Run(); err != nil {
		ui.Info("Could not start ssh-agent service automatically")
		fmt.Println("   Please run as Administrator:")
		fmt.Println("   Set-Service -Name ssh-agent -StartupType Automatic")
		fmt.Println("   Start-Service ssh-agent")
		fmt.Println()
	} else {
		ui.Success("ssh-agent service started")
	}

	// Set ssh-agent to automatic startup
	autoCmd := exec.Command("powershell", "-Command", "Set-Service -Name ssh-agent -StartupType Automatic")
	autoCmd.Run() // Ignore errors

	// Add keys to ssh-agent
	fmt.Println()
	fmt.Println("2. Adding SSH keys to agent...")

	addedCount := 0
	for _, user := range cfg.Users {
		if user.SSHKeyPath == "" {
			continue
		}

		fmt.Printf("   Adding key: %s\n", user.SSHKeyPath)

		addCmd := exec.Command("ssh-add", user.SSHKeyPath)
		output, err := addCmd.CombinedOutput()

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to add key for %s", user.Alias))
			fmt.Printf("   Error: %s\n", string(output))
		} else {
			ui.Success(fmt.Sprintf("Added key for %s", user.Alias))
			addedCount++
		}
	}

	fmt.Println()
	fmt.Printf("Added %d SSH keys to agent\n", addedCount)

	// List loaded keys
	fmt.Println()
	fmt.Println("3. Verifying loaded keys...")
	listCmd := exec.Command("ssh-add", "-l")
	output, err := listCmd.Output()
	if err != nil {
		ui.Info("No keys currently loaded in ssh-agent")
	} else {
		fmt.Println(string(output))
	}

	return nil
}

func setupUnixSSH(cfg *config.Config) error {
	fmt.Println("Unix/Linux SSH Setup:")
	fmt.Println()

	// Check if ssh-agent is running
	agentCheck := exec.Command("pgrep", "ssh-agent")
	if err := agentCheck.Run(); err != nil {
		fmt.Println("1. Starting ssh-agent...")
		fmt.Println("   Run: eval $(ssh-agent)")
		fmt.Println()
	} else {
		ui.Success("ssh-agent is running")
		fmt.Println()
	}

	// Add keys
	fmt.Println("2. Adding SSH keys to agent...")

	addedCount := 0
	for _, user := range cfg.Users {
		if user.SSHKeyPath == "" {
			continue
		}

		fmt.Printf("   Adding key: %s\n", user.SSHKeyPath)

		addCmd := exec.Command("ssh-add", user.SSHKeyPath)
		output, err := addCmd.CombinedOutput()

		if err != nil {
			ui.Error(fmt.Sprintf("Failed to add key for %s", user.Alias))
			fmt.Printf("   Error: %s\n", string(output))
		} else {
			ui.Success(fmt.Sprintf("Added key for %s", user.Alias))
			addedCount++
		}
	}

	fmt.Println()
	fmt.Printf("Added %d SSH keys to agent\n", addedCount)

	// List loaded keys
	fmt.Println()
	fmt.Println("3. Verifying loaded keys...")
	listCmd := exec.Command("ssh-add", "-l")
	output, err := listCmd.Output()
	if err != nil {
		ui.Info("No keys currently loaded in ssh-agent")
	} else {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Println("  ", line)
			}
		}
	}

	return nil
}
