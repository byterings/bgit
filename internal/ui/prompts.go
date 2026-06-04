package ui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"golang.org/x/term"
)

var nonTTYSecretReader = bufio.NewReader(os.Stdin)

// PromptUserInfo prompts for user information interactively
func PromptUserInfo() (alias, name, email, githubUsername string, err error) {
	// Prompt for alias
	aliasPrompt := &survey.Input{
		Message: "Alias (e.g., work, personal, freelance):",
		Help:    "Short name for switching identities - use lowercase, no spaces",
	}
	if err := survey.AskOne(aliasPrompt, &alias, survey.WithValidator(survey.Required)); err != nil {
		return "", "", "", "", err
	}

	// Prompt for name
	namePrompt := &survey.Input{
		Message: "Full name:",
		Help:    "Your full name for Git commits (e.g., John Doe)",
	}
	if err := survey.AskOne(namePrompt, &name, survey.WithValidator(survey.Required)); err != nil {
		return "", "", "", "", err
	}

	// Prompt for email
	emailPrompt := &survey.Input{
		Message: "Email address:",
		Help:    "Your email for Git commits (e.g., john@example.com)",
	}
	emailValidator := func(val interface{}) error {
		if str, ok := val.(string); ok {
			if !isValidEmail(str) {
				return fmt.Errorf("invalid email format")
			}
		}
		return nil
	}
	if err := survey.AskOne(emailPrompt, &email, survey.WithValidator(survey.Required), survey.WithValidator(emailValidator)); err != nil {
		return "", "", "", "", err
	}

	// Prompt for GitHub username
	githubPrompt := &survey.Input{
		Message: "GitHub username:",
		Help:    "Your GitHub username (e.g., johndoe)",
	}
	if err := survey.AskOne(githubPrompt, &githubUsername, survey.WithValidator(survey.Required)); err != nil {
		return "", "", "", "", err
	}

	return alias, name, email, githubUsername, nil
}

// PromptSSHKeyOption prompts for SSH key setup option
func PromptSSHKeyOption() (string, error) {
	var choice string
	prompt := &survey.Select{
		Message: "How do you want to set up SSH key?",
		Options: []string{
			"Generate new key pair (Recommended)",
			"Import existing key",
			"Skip for now (add manually later)",
		},
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return "", err
	}
	return choice, nil
}

// PromptExistingKeyPath prompts for existing SSH key path
func PromptExistingKeyPath() (string, error) {
	var path string
	prompt := &survey.Input{
		Message: "Path to existing SSH private key:",
		Help:    "Full path to your private key file (e.g., ~/.ssh/id_ed25519)",
	}
	if err := survey.AskOne(prompt, &path, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return path, nil
}

// PromptConfirmation prompts for yes/no confirmation
func PromptConfirmation(message string) (bool, error) {
	var confirmed bool
	prompt := &survey.Confirm{
		Message: message,
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirmed); err != nil {
		return false, err
	}
	return confirmed, nil
}

// PromptPasswordConfirmation prompts for a password and matching confirmation.
func PromptPasswordConfirmation(message, confirmMessage string) (string, error) {
	password, err := promptSecret(message)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	confirmation, err := promptSecret(confirmMessage)
	if err != nil {
		return "", err
	}
	if password != confirmation {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

// PromptPassword prompts for a single password value.
func PromptPassword(message string) (string, error) {
	password, err := promptSecret(message)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	return password, nil
}

func promptSecret(message string) (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		var value string
		prompt := &survey.Password{Message: message}
		if err := survey.AskOne(prompt, &value, survey.WithValidator(survey.Required)); err != nil {
			return "", err
		}
		return value, nil
	}

	fmt.Fprint(os.Stderr, message+" ")
	value, err := nonTTYSecretReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(value, "\r\n"), nil
}

// isValidEmail checks if email format is valid
func isValidEmail(email string) bool {
	// Simple email validation regex
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
