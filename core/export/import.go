package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
	coressh "github.com/byterings/bgit/core/ssh"
	"github.com/byterings/bgit/internal/platform"
)

// ImportArchive decrypts a .bgit archive and restores its validated config.
func ImportArchive(path string, password string) (*models.ImportArchiveResult, error) {
	archiveBytes, _, err := decryptArchiveFile(path, password)
	if err != nil {
		return nil, err
	}

	contents, err := readPayloadContents(archiveBytes)
	if err != nil {
		return nil, err
	}

	var imported config.Config
	if err := toml.Unmarshal(contents.ConfigData, &imported); err != nil {
		return nil, fmt.Errorf("failed to decode imported config: %w", err)
	}
	if err := restoreArchivedKeys(&imported, contents.Keys); err != nil {
		return nil, err
	}
	if err := config.ValidateConfig(&imported); err != nil {
		return nil, fmt.Errorf("invalid imported config: %w", err)
	}

	if err := config.SaveConfig(&imported); err != nil {
		return nil, fmt.Errorf("failed to save imported config: %w", err)
	}
	if err := coressh.UpdateSSHConfig(imported.Users); err != nil {
		return nil, fmt.Errorf("failed to regenerate SSH config: %w", err)
	}

	return &models.ImportArchiveResult{
		UsersCount:      len(imported.Users),
		WorkspacesCount: len(imported.Workspaces),
		BindingsCount:   len(imported.Bindings),
		ActiveUser:      imported.ActiveUser,
	}, nil
}

func restoreArchivedKeys(cfg *config.Config, keys map[string]payloadKeyPair) error {
	if len(keys) == 0 {
		return nil
	}

	sshDir, err := platform.GetSSHDir()
	if err != nil {
		return err
	}
	if err := platform.MkdirSecure(sshDir); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	for i := range cfg.Users {
		user := &cfg.Users[i]
		pair, ok := keys[user.Alias]
		if !ok {
			continue
		}
		if len(pair.PrivateKey) == 0 || len(pair.PublicKey) == 0 {
			return fmt.Errorf("archive is missing complete SSH key pair for %s", user.Alias)
		}
		if err := validateKeyAlias(user.Alias); err != nil {
			return err
		}

		privatePath := filepath.Join(sshDir, "bgit_"+user.Alias)
		publicPath := privatePath + ".pub"

		if err := os.WriteFile(privatePath, pair.PrivateKey, 0600); err != nil {
			return fmt.Errorf("failed to restore private key for %s: %w", user.Alias, err)
		}
		if err := os.Chmod(privatePath, 0600); err != nil {
			return fmt.Errorf("failed to secure private key for %s: %w", user.Alias, err)
		}
		if err := os.WriteFile(publicPath, pair.PublicKey, 0644); err != nil {
			return fmt.Errorf("failed to restore public key for %s: %w", user.Alias, err)
		}
		if err := os.Chmod(publicPath, 0644); err != nil {
			return fmt.Errorf("failed to set public key permissions for %s: %w", user.Alias, err)
		}

		user.SSHKeyPath = privatePath
	}

	return nil
}
