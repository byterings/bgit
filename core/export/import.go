package export

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
)

// ImportArchive decrypts a .bgit archive and restores its validated config.
func ImportArchive(path string, password string) (*models.ImportArchiveResult, error) {
	archiveBytes, _, err := decryptArchiveFile(path, password)
	if err != nil {
		return nil, err
	}

	configData, err := readPayloadConfig(archiveBytes)
	if err != nil {
		return nil, err
	}

	var imported config.Config
	if err := toml.Unmarshal(configData, &imported); err != nil {
		return nil, fmt.Errorf("failed to decode imported config: %w", err)
	}
	if err := config.ValidateConfig(&imported); err != nil {
		return nil, fmt.Errorf("invalid imported config: %w", err)
	}

	if err := config.SaveConfig(&imported); err != nil {
		return nil, fmt.Errorf("failed to save imported config: %w", err)
	}

	return &models.ImportArchiveResult{
		UsersCount:      len(imported.Users),
		WorkspacesCount: len(imported.Workspaces),
		BindingsCount:   len(imported.Bindings),
		ActiveUser:      imported.ActiveUser,
	}, nil
}
