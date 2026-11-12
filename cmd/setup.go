package cmd

import (
	"fmt"

	"github.com/byterings/bgit/core/config"
	coresetup "github.com/byterings/bgit/core/setup"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-time setup for bgit",
	Long: `Run one-time setup for bgit.

This command initializes bgit configuration, installs global pre-push safety checks,
and prepares SSH configuration management.`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	cfg, newlyInitialized, err := ensureConfigInitialized()
	if err != nil {
		return err
	}

	if err := applySetup(cfg, false); err != nil {
		return err
	}

	if newlyInitialized {
		ui.Success("Initialized bgit configuration")
	}
	ui.Success("Setup complete")
	ui.Info("Pre-push safety checks are now enabled globally")
	ui.Info("Optional shell prompt integration: use `bgit prompt --plain` in your shell prompt")
	ui.Info("Next step: run 'bgit add' to add your first identity")
	return nil
}

func ensureConfigInitialized() (*config.Config, bool, error) {
	return coresetup.EnsureConfigInitialized()
}

// applySetup runs the shared setup flow and prints the CLI messages for it.
func applySetup(cfg *config.Config, silent bool) error {
	result, err := coresetup.Apply(cfg, coresetup.Options{})
	if result != nil && result.ReplacedHooksPath != "" && !silent {
		ui.Warning(fmt.Sprintf("Replacing existing global hooks path: %s", result.ReplacedHooksPath))
	}
	if err != nil {
		return err
	}

	if !silent {
		ui.Success(fmt.Sprintf("Installed pre-push hook at %s", result.HookPath))
		ui.Success(fmt.Sprintf("Configured global hooks path: %s", result.HooksDir))
		ui.Success("Updated SSH managed section")
	}

	return nil
}
