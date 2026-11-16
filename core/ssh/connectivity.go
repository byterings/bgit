package ssh

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
	"github.com/byterings/bgit/internal/platform"
)

type ConnectivityResult = models.ConnectivityResult
type ConnectivityCheckResult = models.ConnectivityCheckResult

// CheckGitHubConnectivity tests GitHub SSH auth for each configured user with a key.
func CheckGitHubConnectivity(users []config.User) *ConnectivityCheckResult {
	results := make([]ConnectivityResult, 0, len(users))

	for _, user := range users {
		if user.SSHKeyPath == "" {
			continue
		}

		host := GetHostForUser(user.GitHubUsername)
		cmd := platform.PrepareCommand(exec.Command("ssh", "-T", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=10", fmt.Sprintf("git@%s", host)))
		output, _ := cmd.CombinedOutput()

		results = append(results, classifyConnectivity(user, string(output)))
	}

	return &ConnectivityCheckResult{Results: results}
}

// TestConnection checks GitHub SSH auth for one user without ever prompting.
// It is meant for windowed apps, where a passphrase or host-key prompt would
// hang with nobody to answer it. The managed SSH config must already contain
// the user's host alias.
func TestConnection(user config.User) ConnectivityResult {
	if user.SSHKeyPath == "" {
		return ConnectivityResult{
			Alias:    user.Alias,
			Username: user.GitHubUsername,
			Message:  fmt.Sprintf("%s: no SSH key configured", user.Alias),
			Fix:      "Generate or choose an SSH key for this identity",
		}
	}

	host := GetHostForUser(user.GitHubUsername)
	cmd := platform.PrepareCommand(exec.Command("ssh", "-T",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		fmt.Sprintf("git@%s", host)))
	output, _ := cmd.CombinedOutput()

	result := classifyConnectivity(user, string(output))
	result.Authenticated = authenticatedLogin(string(output))
	if !result.Passed {
		refineConnectionFailure(&result, user, string(output))
	}
	if result.Passed && result.Authenticated != "" && !strings.EqualFold(result.Authenticated, user.GitHubUsername) {
		// The key works, but it belongs to another GitHub account.
		result.Passed = false
		result.Message = fmt.Sprintf("%s: key is registered to GitHub user '%s', expected '%s'", user.Alias, result.Authenticated, user.GitHubUsername)
		result.Fix = "Add this SSH key to the correct GitHub account, or correct the GitHub username"
	}
	return result
}

// refineConnectionFailure recognises failures that the basic classification
// reports as unknown, so a windowed app can explain them without showing raw
// SSH output.
func refineConnectionFailure(result *ConnectivityResult, user config.User, output string) {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "could not resolve hostname"),
		strings.Contains(lower, "network is unreachable"),
		strings.Contains(lower, "no route to host"),
		strings.Contains(lower, "operation timed out"):
		result.Message = fmt.Sprintf("%s: connection failed", user.Alias)
	case strings.Contains(lower, "host key verification failed"):
		result.Message = fmt.Sprintf("%s: host key verification failed", user.Alias)
		result.Fix = "The saved GitHub host key does not match. Check ~/.ssh/known_hosts"
	case strings.Contains(lower, "passphrase"), strings.Contains(lower, "incorrect passphrase"):
		result.Message = fmt.Sprintf("%s: key needs a passphrase", user.Alias)
		result.Fix = "Load the key into the SSH agent with ssh-add, then test again"
	case strings.Contains(lower, "bad permissions"), strings.Contains(lower, "unprotected private key"):
		result.Message = fmt.Sprintf("%s: key file permissions are too open", user.Alias)
		result.Fix = "Run Doctor to see how to fix the key file permissions"
	}
	result.Detail = lastLine(output)
}

func lastLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// classifyConnectivity turns the output of `ssh -T git@<host>` into a result.
func classifyConnectivity(user config.User, outputStr string) ConnectivityResult {
	result := ConnectivityResult{
		Alias:    user.Alias,
		Username: user.GitHubUsername,
	}

	switch {
	case strings.Contains(outputStr, "successfully authenticated"), strings.Contains(outputStr, "Hi "):
		result.Passed = true
		result.Message = fmt.Sprintf("%s: authenticated as %s", user.Alias, user.GitHubUsername)
	case strings.Contains(outputStr, "Permission denied"):
		result.Message = fmt.Sprintf("%s: permission denied", user.Alias)
		result.Fix = "Check SSH key is added to GitHub account"
	case strings.Contains(outputStr, "Connection refused"), strings.Contains(outputStr, "Connection timed out"):
		result.Message = fmt.Sprintf("%s: connection failed", user.Alias)
	default:
		result.Message = fmt.Sprintf("%s: unknown response", user.Alias)
	}

	return result
}

// authenticatedLogin extracts the login from GitHub's greeting,
// "Hi <login>! You've successfully authenticated".
func authenticatedLogin(output string) string {
	_, rest, found := strings.Cut(output, "Hi ")
	if !found {
		return ""
	}
	login, _, found := strings.Cut(rest, "!")
	if !found {
		return ""
	}
	return strings.TrimSpace(login)
}
