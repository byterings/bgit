package ui

import (
	"fmt"

	"github.com/byterings/bgit/internal/config"
)

// PrintUsersList prints the list of users in a formatted way
func PrintUsersList(users []config.User, activeUser string) {
	if len(users) == 0 {
		fmt.Println("No users configured yet.")
		fmt.Println("\nAdd your first user with: bgit add")
		return
	}

	fmt.Println("\nConfigured users:")
	fmt.Println()

	for _, user := range users {
		indicator := " "
		if user.Alias == activeUser {
			indicator = "→"
		}

		fmt.Printf("%s %-20s %-30s %s\n",
			indicator,
			user.Alias,
			user.Email,
			user.Name,
		)
	}

	fmt.Println()
	if activeUser == "" {
		fmt.Println("No active user set. Use 'bgit use <alias>' to set one.")
	}
}

// Success prints a success message with checkmark
func Success(message string) {
	fmt.Printf("✓ %s\n", message)
}

// Error prints an error message
func Error(message string) {
	fmt.Printf("✗ %s\n", message)
}

// Info prints an info message
func Info(message string) {
	fmt.Printf("ℹ %s\n", message)
}

// Warning prints a warning message
func Warning(message string) {
	fmt.Printf("⚠ %s\n", message)
}
