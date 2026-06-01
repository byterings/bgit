package ssh

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
)

type ConnectivityResult = models.ConnectivityResult

// CheckGitHubConnectivity tests GitHub SSH auth for each configured user with a key.
func CheckGitHubConnectivity(users []config.User) []ConnectivityResult {
	results := make([]ConnectivityResult, 0, len(users))

	for _, user := range users {
		if user.SSHKeyPath == "" {
			continue
		}

		host := GetHostForUser(user.GitHubUsername)
		cmd := exec.Command("ssh", "-T", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=10", fmt.Sprintf("git@%s", host))
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)

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

		results = append(results, result)
	}

	return results
}
