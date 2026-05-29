package cmd

import (
	"fmt"

	coreidentity "github.com/byterings/bgit/core/identity"
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
	cfg, exists, err := loadConfigReadOnly()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if !exists {
		printPromptNone()
		return nil
	}

	resolution, err := coreidentity.GetEffectiveResolution(cfg)
	if err != nil || resolution == nil || resolution.User == nil {
		printPromptNone()
		return nil
	}

	if promptPlain {
		fmt.Print(resolution.Alias)
		return nil
	}

	fmt.Printf("[bgit:%s]", resolution.Alias)
	return nil
}

func printPromptNone() {
	if promptPlain {
		fmt.Print("none")
	} else {
		fmt.Print("[bgit:none]")
	}
}
