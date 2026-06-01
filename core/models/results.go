package models

// ConnectivityResult reports an SSH authentication check against GitHub.
type ConnectivityResult struct {
	Passed   bool
	Alias    string
	Username string
	Fix      string
	Message  string
}

// SSHAgentSetupReport describes key loading results for setup flows.
type SSHAgentSetupReport struct {
	Added  map[string]string
	Failed map[string]string
}
