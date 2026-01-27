package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/byterings/bgit/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize bgit configuration",
	Long:  `Initialize bgit by creating the configuration directory. This is optional - bgit will auto-initialize on first use.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if already initialized
	exists, err := config.ConfigExists()
	if err != nil {
		return fmt.Errorf("failed to check config: %w", err)
	}

	if exists {
		configDir, _ := config.GetConfigDir()
		fmt.Printf("bgit is already initialized at: %s\n", configDir)
		return nil
	}

	// Create config directory
	if err := config.CreateConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create backup directory
	if err := config.CreateBackupDir(); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create empty config
	cfg := config.NewConfig()
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	configDir, _ := config.GetConfigDir()
	fmt.Printf("✓ bgit initialized at: %s\n", configDir)
	fmt.Println("\nNext: bgit add user")

	return nil
}
