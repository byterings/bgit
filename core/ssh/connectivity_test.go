package ssh

import (
	"testing"

	"github.com/byterings/bgit/core/config"
)

func TestClassifyConnectivity(t *testing.T) {
	user := config.User{Alias: "work", GitHubUsername: "work-gh"}

	cases := []struct {
		name   string
		output string
		passed bool
		fix    bool
	}{
		{"authenticated", "Hi work-gh! You've successfully authenticated, but GitHub does not provide shell access.", true, false},
		{"permission denied", "git@github.com: Permission denied (publickey).", false, true},
		{"timeout", "ssh: connect to host github.com port 22: Connection timed out", false, false},
		{"unknown", "something unexpected", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifyConnectivity(user, tc.output)
			if result.Passed != tc.passed {
				t.Fatalf("Passed = %v, want %v", result.Passed, tc.passed)
			}
			if (result.Fix != "") != tc.fix {
				t.Fatalf("Fix = %q, want set=%v", result.Fix, tc.fix)
			}
		})
	}
}

func TestAuthenticatedLogin(t *testing.T) {
	if got := authenticatedLogin("Hi octo-cat! You've successfully authenticated"); got != "octo-cat" {
		t.Fatalf("login = %q", got)
	}
	if got := authenticatedLogin("Permission denied (publickey)."); got != "" {
		t.Fatalf("login should be empty, got %q", got)
	}
}

func TestTestConnectionWithoutKey(t *testing.T) {
	result := TestConnection(config.User{Alias: "work", GitHubUsername: "work-gh"})
	if result.Passed || result.Fix == "" {
		t.Fatalf("expected a failed result with a fix, got %+v", result)
	}
}

func TestRefineConnectionFailure(t *testing.T) {
	user := config.User{Alias: "work", GitHubUsername: "work-gh"}

	cases := []struct {
		output  string
		message string
		fix     bool
	}{
		{"ssh: Could not resolve hostname github.com-work-gh: Name or service not known", "work: connection failed", false},
		{"Host key verification failed.", "work: host key verification failed", true},
		{"Load key \"/k\": bad permissions\ngit@github.com: Permission denied (publickey).", "work: key file permissions are too open", true},
	}
	for _, tc := range cases {
		result := classifyConnectivity(user, tc.output)
		refineConnectionFailure(&result, user, tc.output)
		if result.Message != tc.message {
			t.Fatalf("output %q: message = %q, want %q", tc.output, result.Message, tc.message)
		}
		if (result.Fix != "") != tc.fix {
			t.Fatalf("output %q: fix = %q", tc.output, result.Fix)
		}
		if result.Detail == "" {
			t.Fatalf("output %q: detail should be set", tc.output)
		}
	}
}
