package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/platform"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	doctorNetwork bool
	doctorFix     bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration issues",
	Long: `Check bgit configuration health and diagnose common issues.

Runs checks on:
- Config file validity
- SSH key existence and permissions
- SSH config entries
- SSH agent status
- Git config alignment

Examples:
  bgit doctor              # Run basic diagnostics
  bgit doctor --network    # Include GitHub connectivity tests
  bgit doctor --fix        # Auto-fix permission issues`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().BoolVarP(&doctorNetwork, "network", "n", false, "Test GitHub SSH connectivity")
	doctorCmd.Flags().BoolVarP(&doctorFix, "fix", "f", false, "Auto-fix permission issues")
}

type checkResult struct {
	passed  bool
	message string
	fix     string // Suggested fix command
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("Checking bgit configuration...")
	fmt.Println()

	errors := 0
	warnings := 0
	fixed := 0

	// 1. Config checks
	fmt.Println("Config")
	fmt.Println("──────")

	configResults := checkConfig()
	for _, r := range configResults {
		printCheckResult(r)
		if !r.passed {
			errors++
		}
	}

	// Load config for remaining checks
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println()
		ui.Error(fmt.Sprintf("Cannot continue: %v", err))
		return nil
	}

	// 2. SSH directory and file checks
	fmt.Println()
	fmt.Println("SSH Setup")
	fmt.Println("─────────")

	sshResults, sshFixed := checkSSH(cfg, doctorFix)
	for _, r := range sshResults {
		printCheckResult(r)
		if !r.passed && r.fix == "" {
			errors++
		} else if !r.passed {
			warnings++
		}
	}
	fixed += sshFixed

	// 3. SSH agent checks
	fmt.Println()
	fmt.Println("SSH Agent")
	fmt.Println("─────────")

	agentResults := checkSSHAgent()
	for _, r := range agentResults {
		printCheckResult(r)
		if !r.passed && r.fix == "" {
			errors++
		} else if !r.passed {
			warnings++
		}
	}

	// 4. Git config checks
	fmt.Println()
	fmt.Println("Git Config")
	fmt.Println("──────────")

	gitResults := checkGitConfig(cfg)
	for _, r := range gitResults {
		printCheckResult(r)
		if !r.passed && r.fix == "" {
			errors++
		} else if !r.passed {
			warnings++
		}
	}

	// 5. Network checks (optional)
	if doctorNetwork {
		fmt.Println()
		fmt.Println("GitHub Connectivity")
		fmt.Println("───────────────────")

		netResults := checkGitHubConnectivity(cfg)
		for _, r := range netResults {
			printCheckResult(r)
			if !r.passed {
				errors++
			}
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("─────────")

	if fixed > 0 {
		ui.Success(fmt.Sprintf("Auto-fixed %d issue(s)", fixed))
	}

	if errors == 0 && warnings == 0 {
		ui.Success("All checks passed!")
	} else if errors == 0 {
		ui.Warning(fmt.Sprintf("%d warning(s)", warnings))
	} else {
		ui.Error(fmt.Sprintf("%d error(s), %d warning(s)", errors, warnings))
	}

	return nil
}

func printCheckResult(r checkResult) {
	if r.passed {
		fmt.Printf("  ✓ %s\n", r.message)
	} else if r.fix != "" {
		fmt.Printf("  ⚠ %s\n", r.message)
		fmt.Printf("    → %s\n", r.fix)
	} else {
		fmt.Printf("  ✗ %s\n", r.message)
	}
}

func checkConfig() []checkResult {
	var results []checkResult

	// Check if config exists
	exists, err := config.ConfigExists()
	if err != nil {
		results = append(results, checkResult{
			passed:  false,
			message: fmt.Sprintf("Error checking config: %v", err),
		})
		return results
	}

	if !exists {
		results = append(results, checkResult{
			passed:  false,
			message: "Config file not found",
			fix:     "Run: bgit add",
		})
		return results
	}

	results = append(results, checkResult{
		passed:  true,
		message: "Config file exists",
	})

	// Try to load config
	cfg, err := config.LoadConfig()
	if err != nil {
		results = append(results, checkResult{
			passed:  false,
			message: fmt.Sprintf("Config file invalid: %v", err),
		})
		return results
	}

	results = append(results, checkResult{
		passed:  true,
		message: "Config file valid",
	})

	// Check users configured
	if len(cfg.Users) == 0 {
		results = append(results, checkResult{
			passed:  false,
			message: "No users configured",
			fix:     "Run: bgit add",
		})
	} else {
		results = append(results, checkResult{
			passed:  true,
			message: fmt.Sprintf("%d user(s) configured", len(cfg.Users)),
		})
	}

	// Check active user
	if cfg.ActiveUser == "" {
		results = append(results, checkResult{
			passed:  false,
			message: "No active user set",
			fix:     "Run: bgit use <alias>",
		})
	} else {
		user := cfg.FindUserByAlias(cfg.ActiveUser)
		if user == nil {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("Active user '%s' not found in config", cfg.ActiveUser),
			})
		} else {
			results = append(results, checkResult{
				passed:  true,
				message: fmt.Sprintf("Active user: %s", cfg.ActiveUser),
			})
		}
	}

	return results
}

func checkSSH(cfg *config.Config, autoFix bool) ([]checkResult, int) {
	var results []checkResult
	fixed := 0

	// Check .ssh directory
	sshDir, err := platform.GetSSHDir()
	if err != nil {
		results = append(results, checkResult{
			passed:  false,
			message: fmt.Sprintf("Cannot determine SSH directory: %v", err),
		})
		return results, fixed
	}

	// Check if .ssh exists
	info, err := os.Stat(sshDir)
	if os.IsNotExist(err) {
		results = append(results, checkResult{
			passed:  false,
			message: "SSH directory does not exist",
			fix:     fmt.Sprintf("Run: mkdir -p %s && chmod 700 %s", sshDir, sshDir),
		})
		return results, fixed
	}

	// Check .ssh permissions (Unix only)
	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		if mode != 0700 {
			if autoFix {
				if err := os.Chmod(sshDir, 0700); err == nil {
					results = append(results, checkResult{
						passed:  true,
						message: "SSH directory permissions fixed (700)",
					})
					fixed++
				} else {
					results = append(results, checkResult{
						passed:  false,
						message: fmt.Sprintf("SSH directory has wrong permissions (%o)", mode),
						fix:     fmt.Sprintf("chmod 700 %s", sshDir),
					})
				}
			} else {
				results = append(results, checkResult{
					passed:  false,
					message: fmt.Sprintf("SSH directory has wrong permissions (%o, should be 700)", mode),
					fix:     fmt.Sprintf("chmod 700 %s", sshDir),
				})
			}
		} else {
			results = append(results, checkResult{
				passed:  true,
				message: "SSH directory permissions OK (700)",
			})
		}
	}

	// Check SSH keys for each user
	for _, user := range cfg.Users {
		if user.SSHKeyPath == "" {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("No SSH key path for '%s'", user.Alias),
				fix:     fmt.Sprintf("Run: bgit update %s", user.Alias),
			})
			continue
		}

		keyPath := user.SSHKeyPath

		// Check key exists
		keyInfo, err := os.Stat(keyPath)
		if os.IsNotExist(err) {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("SSH key missing for '%s': %s", user.Alias, keyPath),
				fix:     fmt.Sprintf("Run: bgit update %s --generate-key", user.Alias),
			})
			continue
		}

		// Check key permissions (Unix only)
		if runtime.GOOS != "windows" {
			mode := keyInfo.Mode().Perm()
			if mode != 0600 {
				if autoFix {
					if err := os.Chmod(keyPath, 0600); err == nil {
						results = append(results, checkResult{
							passed:  true,
							message: fmt.Sprintf("SSH key '%s' permissions fixed (600)", user.Alias),
						})
						fixed++
					} else {
						results = append(results, checkResult{
							passed:  false,
							message: fmt.Sprintf("SSH key '%s' has wrong permissions (%o)", user.Alias, mode),
							fix:     fmt.Sprintf("chmod 600 %s", keyPath),
						})
					}
				} else {
					results = append(results, checkResult{
						passed:  false,
						message: fmt.Sprintf("SSH key '%s' has wrong permissions (%o, should be 600)", user.Alias, mode),
						fix:     fmt.Sprintf("chmod 600 %s", keyPath),
					})
				}
			} else {
				results = append(results, checkResult{
					passed:  true,
					message: fmt.Sprintf("SSH key '%s' exists with correct permissions", user.Alias),
				})
			}
		} else {
			results = append(results, checkResult{
				passed:  true,
				message: fmt.Sprintf("SSH key '%s' exists", user.Alias),
			})
		}
	}

	// Check SSH config
	sshConfigPath, _ := platform.GetSSHConfigPath()
	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		results = append(results, checkResult{
			passed:  false,
			message: "SSH config file not found",
			fix:     "Run: bgit sync --fix",
		})
	} else {
		// Read and check for bgit entries
		content, err := os.ReadFile(sshConfigPath)
		if err == nil {
			if strings.Contains(string(content), "BEGIN BRGIT MANAGED") {
				results = append(results, checkResult{
					passed:  true,
					message: "SSH config has bgit entries",
				})
			} else {
				results = append(results, checkResult{
					passed:  false,
					message: "SSH config missing bgit entries",
					fix:     "Run: bgit sync --fix",
				})
			}
		}
	}

	return results, fixed
}

func checkSSHAgent() []checkResult {
	var results []checkResult

	// Check if SSH agent is running
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		results = append(results, checkResult{
			passed:  false,
			message: "SSH agent not running (SSH_AUTH_SOCK not set)",
			fix:     "Run: eval $(ssh-agent)",
		})
		return results
	}

	// Verify socket exists
	if _, err := os.Stat(authSock); os.IsNotExist(err) {
		results = append(results, checkResult{
			passed:  false,
			message: "SSH agent socket missing",
			fix:     "Run: eval $(ssh-agent)",
		})
		return results
	}

	results = append(results, checkResult{
		passed:  true,
		message: "SSH agent running",
	})

	// Try to list keys
	cmd := exec.Command("ssh-add", "-l")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "no identities") {
			results = append(results, checkResult{
				passed:  false,
				message: "No keys loaded in SSH agent",
				fix:     "Run: ssh-add ~/.ssh/bgit_*",
			})
		} else {
			results = append(results, checkResult{
				passed:  false,
				message: "Could not list SSH agent keys",
			})
		}
	} else {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		results = append(results, checkResult{
			passed:  true,
			message: fmt.Sprintf("%d key(s) loaded in agent", len(lines)),
		})
	}

	return results
}

func checkGitConfig(cfg *config.Config) []checkResult {
	var results []checkResult

	if cfg.ActiveUser == "" {
		return results
	}

	user := cfg.FindUserByAlias(cfg.ActiveUser)
	if user == nil {
		return results
	}

	// Check git user.name
	cmd := exec.Command("git", "config", "--global", "user.name")
	output, err := cmd.Output()
	if err != nil {
		results = append(results, checkResult{
			passed:  false,
			message: "Could not read git user.name",
		})
	} else {
		name := strings.TrimSpace(string(output))
		if name == user.Name {
			results = append(results, checkResult{
				passed:  true,
				message: fmt.Sprintf("user.name = %s", name),
			})
		} else {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("user.name mismatch: '%s' (expected: '%s')", name, user.Name),
				fix:     "Run: bgit sync --fix",
			})
		}
	}

	// Check git user.email
	cmd = exec.Command("git", "config", "--global", "user.email")
	output, err = cmd.Output()
	if err != nil {
		results = append(results, checkResult{
			passed:  false,
			message: "Could not read git user.email",
		})
	} else {
		email := strings.TrimSpace(string(output))
		if email == user.Email {
			results = append(results, checkResult{
				passed:  true,
				message: fmt.Sprintf("user.email = %s", email),
			})
		} else {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("user.email mismatch: '%s' (expected: '%s')", email, user.Email),
				fix:     "Run: bgit sync --fix",
			})
		}
	}

	return results
}

func checkGitHubConnectivity(cfg *config.Config) []checkResult {
	var results []checkResult

	for _, user := range cfg.Users {
		if user.SSHKeyPath == "" {
			continue
		}

		// Test SSH connection to GitHub with this identity
		host := fmt.Sprintf("github.com-%s", user.GitHubUsername)

		cmd := exec.Command("ssh", "-T", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=10", fmt.Sprintf("git@%s", host))
		output, _ := cmd.CombinedOutput()

		// GitHub returns exit code 1 even on success, check output
		outputStr := string(output)
		if strings.Contains(outputStr, "successfully authenticated") || strings.Contains(outputStr, "Hi ") {
			results = append(results, checkResult{
				passed:  true,
				message: fmt.Sprintf("%s: authenticated as %s", user.Alias, user.GitHubUsername),
			})
		} else if strings.Contains(outputStr, "Permission denied") {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("%s: permission denied", user.Alias),
				fix:     "Check SSH key is added to GitHub account",
			})
		} else if strings.Contains(outputStr, "Connection refused") || strings.Contains(outputStr, "Connection timed out") {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("%s: connection failed", user.Alias),
			})
		} else {
			results = append(results, checkResult{
				passed:  false,
				message: fmt.Sprintf("%s: unknown response", user.Alias),
			})
		}
	}

	return results
}
