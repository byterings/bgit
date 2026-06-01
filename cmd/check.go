package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/byterings/bgit/core/config"
	coreidentity "github.com/byterings/bgit/core/identity"
	corerepo "github.com/byterings/bgit/core/repo"
	"github.com/byterings/bgit/internal/git"
	"github.com/byterings/bgit/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var checkFromHook bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run pre-push safety checks",
	Long: `Validate that repository owner, active user, git config, and remote settings
are aligned before push operations.`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().BoolVar(&checkFromHook, "from-hook", false, "Internal flag for pre-push hook")
	checkCmd.Flags().MarkHidden("from-hook")
}

func runCheck(cmd *cobra.Command, args []string) error {
	if err := autoInit(); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine current directory: %w", err)
	}

	repoRoot := coreidentity.FindGitRoot(cwd)
	if repoRoot == "" {
		if checkFromHook {
			return fmt.Errorf("not in a git repository")
		}
		ui.Info("Not inside a git repository; nothing to check")
		return nil
	}

	remoteURL, _ := getRemoteURL("origin")
	ownerResult := corerepo.ResolveRepoOwner(cfg, repoRoot, remoteURL)
	ownerAlias, ownerSource := "", ""
	var ownerUser *config.User
	if ownerResult != nil && ownerResult.Owner != nil {
		ownerAlias, ownerSource, ownerUser = ownerResult.Owner.Alias, ownerResult.Owner.Source, ownerResult.Owner.User
	}

	if cfg.ActiveUser == "" && ownerAlias == "" {
		return fmt.Errorf("no active user set. Run: bgit use <alias>")
	}

	expectedAlias := ownerAlias
	expectedSource := ownerSource
	expectedUser := ownerUser
	if expectedAlias == "" {
		expectedAlias = cfg.ActiveUser
		expectedSource = "active"
		expectedUser = cfg.FindUserByAlias(expectedAlias)
	}

	if expectedUser == nil {
		return fmt.Errorf("expected user '%s' not found in config", expectedAlias)
	}

	if !checkFromHook {
		fmt.Printf("Repository: %s\n", repoRoot)
		fmt.Printf("Repo owner: %s", expectedAlias)
		if expectedSource != "" {
			fmt.Printf(" (%s)", expectedSource)
		}
		fmt.Println()
		if cfg.ActiveUser != "" {
			fmt.Printf("Active user: %s\n", cfg.ActiveUser)
		}
		fmt.Println()
	}

	useActiveForValidation := false
	if ownerAlias != "" && cfg.ActiveUser != "" && ownerAlias != cfg.ActiveUser {
		confirmedMismatch, err := handleActiveOwnerMismatch(cfg.ActiveUser, ownerAlias)
		if err != nil {
			return err
		}
		useActiveForValidation = confirmedMismatch
	}

	if useActiveForValidation {
		expectedAlias = cfg.ActiveUser
		expectedSource = "active-confirmed"
		expectedUser = cfg.FindUserByAlias(expectedAlias)
		if expectedUser == nil {
			return fmt.Errorf("active user '%s' not found in config", expectedAlias)
		}

		if !checkFromHook {
			ui.Info(fmt.Sprintf("Proceeding with active user '%s' after confirmation", expectedAlias))
			fmt.Println()
		}
	}

	if err := checkGitConfigMatches(expectedUser, expectedAlias); err != nil {
		return err
	}

	if err := checkOriginRemoteMatches(expectedUser, expectedAlias); err != nil {
		return err
	}

	if !checkFromHook {
		ui.Success("Safety checks passed")
	}

	return nil
}

func handleActiveOwnerMismatch(activeAlias, ownerAlias string) (bool, error) {
	if os.Getenv("BGIT_ALLOW_MISMATCH") == "1" {
		ui.Warning(fmt.Sprintf("Mismatch bypassed by BGIT_ALLOW_MISMATCH=1 (active: %s, owner: %s)", activeAlias, ownerAlias))
		return true, nil
	}

	ui.Warning("Identity mismatch detected")
	fmt.Printf("  Active user: %s\n", activeAlias)
	fmt.Printf("  Repo owner:  %s\n", ownerAlias)

	if !isInteractiveSession() {
		return false, fmt.Errorf("active user '%s' does not match repo owner '%s'. Run: bgit use %s (or set BGIT_ALLOW_MISMATCH=1)", activeAlias, ownerAlias, ownerAlias)
	}

	confirmed, err := ui.PromptConfirmation(fmt.Sprintf("Active user is '%s' and repo owner is '%s'. Continue push?", activeAlias, ownerAlias))
	if err != nil {
		return false, err
	}
	if !confirmed {
		return false, fmt.Errorf("push cancelled")
	}

	ui.Warning("Continuing push with identity mismatch by user confirmation")
	return true, nil
}

func checkGitConfigMatches(expectedUser *config.User, expectedAlias string) error {
	gitName, gitEmail, err := git.GetGlobalUser()
	if err != nil {
		return fmt.Errorf("failed to read git config: %w", err)
	}

	if strings.TrimSpace(gitName) != expectedUser.Name || strings.TrimSpace(gitEmail) != expectedUser.Email {
		return fmt.Errorf("git config mismatch for '%s'. Expected %s <%s>, got %s <%s>. Run: bgit use %s",
			expectedAlias,
			expectedUser.Name,
			expectedUser.Email,
			strings.TrimSpace(gitName),
			strings.TrimSpace(gitEmail),
			expectedAlias,
		)
	}

	return nil
}

func checkOriginRemoteMatches(expectedUser *config.User, expectedAlias string) error {
	remoteURL, err := getRemoteURL("origin")
	if err != nil {
		return fmt.Errorf("failed to read origin remote: %w", err)
	}
	if remoteURL == "" {
		return fmt.Errorf("origin remote is empty")
	}

	expectedURL, err := corerepo.ConvertToBgitURL(remoteURL, expectedUser.GitHubUsername)
	if err != nil {
		return fmt.Errorf("unsupported origin URL format: %w", err)
	}

	if remoteURL == expectedURL {
		return nil
	}

	ui.Warning("Remote URL does not match expected identity")
	fmt.Printf("  Current:  %s\n", remoteURL)
	fmt.Printf("  Expected: %s\n", expectedURL)

	if isInteractiveSession() {
		confirmed, err := ui.PromptConfirmation("Fix remote URL automatically now?")
		if err != nil {
			return err
		}
		if confirmed {
			if err := setRemoteURL("origin", expectedURL); err != nil {
				return fmt.Errorf("failed to auto-fix remote URL: %w", err)
			}
			ui.Success("Remote URL fixed automatically")
			return nil
		}
	}

	return fmt.Errorf("origin remote is not aligned with '%s'. Run: bgit remote fix", expectedAlias)
}

func isInteractiveSession() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}
