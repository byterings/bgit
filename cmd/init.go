package cmd

import (
	"fmt"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:        "init",
	Short:      "Initialize bgit configuration",
	Long:       `Initialize bgit by creating the configuration directory. This is optional - bgit will auto-initialize on first use.`,
	Deprecated: "use 'bgit setup' instead",
	RunE:       runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	ui.Warning("'bgit init' is deprecated. Use 'bgit setup' instead.")

	cfg, newlyInitialized, err := ensureConfigInitialized()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	if err := applySetup(cfg, false); err != nil {
		return err
	}

	configDir, _ := config.GetConfigDir()
	if newlyInitialized {
		ui.Success(fmt.Sprintf("bgit initialized at: %s", configDir))
	} else {
		ui.Info(fmt.Sprintf("bgit already initialized at: %s", configDir))
	}
	ui.Success("Setup workflow completed")
	fmt.Println("\nNext: bgit add")
	return nil
}
