package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/byterings/bgit/cmd"
	coressh "github.com/byterings/bgit/core/ssh"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const githubNewSSHKeyURL = "https://github.com/settings/ssh/new"

// openDesktopExternalURL opens a link in the user's browser. Only a fixed set
// of HTTPS sites is allowed, so page content can never make the app open
// arbitrary programs or URLs.
func openDesktopExternalURL(ctx context.Context, url string) error {
	if ctx == nil {
		return fmt.Errorf("desktop runtime is not ready")
	}
	url = strings.TrimSpace(url)
	if !isAllowedExternalURL(url) {
		return fmt.Errorf("this link cannot be opened from bgit")
	}
	wailsruntime.BrowserOpenURL(ctx, url)
	return nil
}

func isAllowedExternalURL(url string) bool {
	for _, prefix := range []string{"https://github.com/", "https://git-scm.com/", "https://bgitcli.com/"} {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}
	return false
}

// testDesktopIdentityConnection checks that GitHub accepts the identity's SSH
// key and that the key belongs to the expected GitHub account.
func testDesktopIdentityConnection(alias string) (*ConnectionTestResult, error) {
	cfg, err := loadExistingDesktopConfig()
	if err != nil {
		return nil, err
	}

	alias = strings.TrimSpace(alias)
	user := cfg.FindUserByAlias(alias)
	if user == nil {
		return nil, fmt.Errorf("identity '%s' not found", alias)
	}

	// The test goes through the identity's host alias, so the managed SSH
	// config must be current before connecting.
	if err := coressh.UpdateSSHConfig(cfg.Users); err != nil {
		return nil, err
	}

	result := coressh.TestConnection(*user)
	return &ConnectionTestResult{
		Alias:          result.Alias,
		GitHubUsername: result.Username,
		Passed:         result.Passed,
		Authenticated:  result.Authenticated,
		Message:        connectionMessage(result),
		Fix:            result.Fix,
	}, nil
}

func connectionMessage(result coressh.ConnectivityResult) string {
	switch {
	case result.Passed:
		return fmt.Sprintf("GitHub accepted the SSH key for '%s' (%s).", result.Alias, result.Username)
	case strings.Contains(result.Message, "permission denied"):
		return fmt.Sprintf("GitHub rejected the SSH key for '%s'. The key is probably not added to the GitHub account yet.", result.Alias)
	case strings.Contains(result.Message, "connection failed"):
		return "Could not reach GitHub. Check the internet connection and try again."
	case strings.Contains(result.Message, "unknown response"):
		if result.Detail != "" {
			return "GitHub did not confirm the SSH key. SSH reported: " + result.Detail
		}
		return "GitHub did not confirm the SSH key."
	default:
		// Messages built by core start with "<alias>: "; show them as a sentence.
		_, detail, found := strings.Cut(result.Message, ": ")
		if !found || detail == "" {
			detail = result.Message
		}
		return strings.ToUpper(detail[:1]) + detail[1:] + "."
	}
}

func loadAppInfo() *AppInfo {
	return &AppInfo{
		Version:        cmd.Version(),
		HookHelperPath: hookHelperPath(),
	}
}
