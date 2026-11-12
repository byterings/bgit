package main

import (
	"fmt"
	"path/filepath"

	coreuninstall "github.com/byterings/bgit/core/uninstall"
)

// uninstallDesktop removes everything bgit changed on this computer, the same
// way `bgit uninstall` does. The app itself is removed by the system's normal
// uninstaller afterwards.
func uninstallDesktop(request UninstallRequest) (*UninstallResult, error) {
	run := coreuninstall.Run(coreuninstall.Options{
		RemoveKeys: request.RemoveKeys,
		// Repositories cloned with the CLI may use a bgit remote without
		// being bound; they would stop working once the SSH alias is gone.
		ScanHome: true,
	})

	result := &UninstallResult{
		RestoredRepos: emptyIfNil(run.RestoredRepos),
		FailedRepos:   emptyIfNil(run.FailedRepos),
		RemovedKeys:   []string{},
		Warnings:      []string{},
		ConfigRemoved: run.ConfigErr == nil,
	}

	if run.SSHConfigErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("The bgit section could not be removed from your SSH config: %v", run.SSHConfigErr))
	}
	if run.HooksErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("The Git hooks setting could not be restored: %v", run.HooksErr))
	}
	if run.ConfigErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("The bgit settings folder could not be removed: %v", run.ConfigErr))
	}
	for _, repo := range run.FailedRepos {
		result.Warnings = append(result.Warnings, fmt.Sprintf("The remote address of %s could not be restored.", filepath.Base(repo)))
	}

	switch identity := run.GitIdentity; {
	case identity == nil:
		result.GitIdentityNote = "bgit settings could not be read, so your Git name and email were left as they are."
	case identity.HadBackup && identity.NameErr == nil && identity.EmailErr == nil:
		result.GitIdentityNote = "Your Git name and email are back to what they were before bgit."
	case identity.HadBackup:
		result.Warnings = append(result.Warnings, "Your previous Git name and email could not be fully restored.")
	case identity.InspectErr != nil:
		result.Warnings = append(result.Warnings, fmt.Sprintf("Your Git name and email could not be checked: %v", identity.InspectErr))
	case identity.MatchingUser != nil:
		result.GitIdentityNote = fmt.Sprintf("Git still uses the name and email of '%s'. bgit did not record what was set before it, so change them in your Git settings if needed.", identity.MatchingUser.Alias)
	default:
		result.GitIdentityNote = "Your Git name and email were left as they are."
	}

	for _, removal := range run.KeyRemovals {
		if removal.Err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("The key file %s could not be deleted: %v", filepath.Base(removal.Path), removal.Err))
			continue
		}
		result.RemovedKeys = append(result.RemovedKeys, removal.Path)
	}

	return result, nil
}

func emptyIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
