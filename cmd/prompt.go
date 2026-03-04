package cmd

import (
	"fmt"

	"github.com/byterings/bgit/internal/config"
	"github.com/byterings/bgit/internal/identity"
	"github.com/spf13/cobra"
)

var promptPlain bool

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Print effective identity for shell prompt integration",
	Long: `Print the effective identity for the current directory.

Use this command from shell prompt scripts to display active/repository identity.`,
	RunE: runPrompt,
}

func init() {
	rootCmd.AddCommand(promptCmd)
	promptCmd.Flags().BoolVar(&promptPlain, "plain", false, "Print alias only")
}

func runPrompt(cmd *cobra.Command, args []string) error {
	if err := autoInit(); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	resolution, err := identity.GetEffectiveResolution(cfg)
	if err != nil || resolution == nil || resolution.User == nil {
		if promptPlain {
			fmt.Print("none")
		} else {
			fmt.Print("[bgit:none]")
		}
		return nil
	}

	if promptPlain {
		fmt.Print(resolution.Alias)
		return nil
	}

	fmt.Printf("[bgit:%s]", resolution.Alias)
	return nil
}
