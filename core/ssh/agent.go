package ssh

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
)

// StartSSHAgent starts the platform SSH agent if bgit can do so.
func StartSSHAgent() {
	if runtime.GOOS == "windows" {
		exec.Command("powershell", "-Command", "Start-Service ssh-agent").Run()
		exec.Command("powershell", "-Command", "Set-Service -Name ssh-agent -StartupType Automatic").Run()
	}
}

// IsSSHAgentRunning reports whether an SSH agent is available in the current session.
func IsSSHAgentRunning() bool {
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		return false
	}
	if _, err := os.Stat(authSock); err != nil {
		return false
	}
	return true
}

type SSHAgentSetupReport = models.SSHAgentSetupReport
type CommandOutputResult = models.CommandOutputResult
type KeyLoadResult = models.KeyLoadResult

// ListAgentKeys returns the output of ssh-add -l.
func ListAgentKeys() (*CommandOutputResult, error) {
	cmd := exec.Command("ssh-add", "-l")
	output, err := cmd.CombinedOutput()
	return &CommandOutputResult{Output: string(output)}, err
}

// AddKeyToAgent adds the given key path to the SSH agent.
func AddKeyToAgent(path string) (*CommandOutputResult, error) {
	cmd := exec.Command("ssh-add", path)
	output, err := cmd.CombinedOutput()
	return &CommandOutputResult{Output: string(output)}, err
}

// EnsureKeyLoaded starts the SSH agent where possible and loads the user's key if missing.
func EnsureKeyLoaded(user *config.User) KeyLoadResult {
	if user == nil || user.SSHKeyPath == "" {
		return KeyLoadResult{}
	}

	StartSSHAgent()
	output, _ := ListAgentKeys()
	if output != nil && strings.Contains(output.Output, user.SSHKeyPath) {
		return KeyLoadResult{}
	}

	result, err := AddKeyToAgent(user.SSHKeyPath)
	if err == nil {
		return KeyLoadResult{
			Output:  result.Output,
			Changed: true,
		}
	}
	if result == nil {
		return KeyLoadResult{}
	}
	return KeyLoadResult{Output: result.Output}
}

// SetupAgentForUsers attempts to start the SSH agent and load all configured user keys.
func SetupAgentForUsers(users []config.User) SSHAgentSetupReport {
	StartSSHAgent()

	report := SSHAgentSetupReport{
		Added:  make(map[string]string),
		Failed: make(map[string]string),
	}

	for _, user := range users {
		if user.SSHKeyPath == "" {
			continue
		}
		if output, err := AddKeyToAgent(user.SSHKeyPath); err != nil {
			report.Failed[user.Alias] = output.Output
		} else {
			report.Added[user.Alias] = user.SSHKeyPath
		}
	}

	return report
}
