package export

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/byterings/bgit/core/config"
	"github.com/byterings/bgit/core/models"
	"github.com/byterings/bgit/internal/platform"
)

const (
	ArchiveExtension    = ".bgit"
	ArchiveFormatV1     = "1"
	ArchiveLayoutV1     = "1"
	ArchiveNamePrefix   = "bgit-export-"
	ManifestPath        = "manifest.json"
	PayloadDir          = "payload"
	PayloadConfigDir    = "payload/config"
	PayloadKeysDir      = "payload/keys"
	PayloadConfigPath   = "payload/config/config.toml"
	EncryptionPlannedIn = "R-010"
)

type CreateArchiveResult struct {
	Path string
}

// CreateArchive exports the current bgit configuration into a stable archive layout.
func CreateArchive(cfg *config.Config, toolVersion string) (*CreateArchiveResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	backupDir, err := config.GetBackupDir()
	if err != nil {
		return nil, err
	}
	if err := platform.MkdirSecure(backupDir); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	createdAt := time.Now().UTC()
	archivePath := filepath.Join(backupDir, archiveFileName(createdAt))

	configData, err := encodeConfig(cfg)
	if err != nil {
		return nil, err
	}

	manifest := buildManifest(cfg, toolVersion, createdAt)
	if err := writeArchive(archivePath, manifest, configData, createdAt); err != nil {
		return nil, err
	}

	return &CreateArchiveResult{Path: archivePath}, nil
}

func archiveFileName(createdAt time.Time) string {
	return ArchiveNamePrefix + createdAt.Format("20060102T150405Z") + ArchiveExtension
}

func encodeConfig(cfg *config.Config) ([]byte, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return nil, fmt.Errorf("failed to encode config for export: %w", err)
	}
	return buf.Bytes(), nil
}

func buildManifest(cfg *config.Config, toolVersion string, createdAt time.Time) models.ExportManifest {
	identities := make([]models.ExportIdentitySummary, 0, len(cfg.Users))
	for _, user := range cfg.Users {
		identities = append(identities, models.ExportIdentitySummary{
			Alias:          user.Alias,
			GitHubUsername: user.GitHubUsername,
			HasSSHKeyPath:  strings.TrimSpace(user.SSHKeyPath) != "",
		})
	}

	return models.ExportManifest{
		FormatVersion: ArchiveFormatV1,
		BgitVersion:   toolVersion,
		CreatedAt:     createdAt.Format(time.RFC3339),
		Archive: models.ExportArchiveDescriptor{
			LayoutVersion: ArchiveLayoutV1,
			Encryption: models.ExportEncryptionState{
				Status:         "plaintext",
				PlannedVersion: EncryptionPlannedIn,
			},
			Files: []models.ExportArchiveFile{
				{Path: ManifestPath, Kind: "file", Required: true},
				{Path: PayloadDir, Kind: "dir", Required: true},
				{Path: PayloadConfigDir, Kind: "dir", Required: true},
				{Path: PayloadConfigPath, Kind: "file", Required: true},
				{Path: PayloadKeysDir, Kind: "dir", Required: true},
			},
		},
		Identities: identities,
	}
}
