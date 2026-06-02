package export

import (
	"bytes"
	"fmt"
	"os"
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

type archiveKeyFile struct {
	Alias string
	Path  string
	Data  []byte
	Mode  int64
}

// CreateArchive exports the current bgit configuration into a stable archive layout.
func CreateArchive(cfg *config.Config, toolVersion string, password string) (*CreateArchiveResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("password is required")
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
	keyFiles, err := collectKeyFiles(cfg)
	if err != nil {
		return nil, err
	}

	manifest := buildManifest(cfg, toolVersion, createdAt)
	archiveBytes, err := buildArchiveBytes(manifest, configData, keyFiles, createdAt)
	if err != nil {
		return nil, err
	}

	header, ciphertext, err := encryptArchive(archiveBytes, toolVersion, createdAt, password)
	if err != nil {
		return nil, err
	}

	if err := writeEncryptedArchive(archivePath, header, ciphertext); err != nil {
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
		summary := models.ExportIdentitySummary{
			Alias:          user.Alias,
			GitHubUsername: user.GitHubUsername,
			HasSSHKeyPath:  strings.TrimSpace(user.SSHKeyPath) != "",
		}
		if summary.HasSSHKeyPath {
			summary.PrivateKeyPath = keyArchivePath(user.Alias, false)
			summary.PublicKeyPath = keyArchivePath(user.Alias, true)
		}
		identities = append(identities, summary)
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

func collectKeyFiles(cfg *config.Config) ([]archiveKeyFile, error) {
	var files []archiveKeyFile
	for _, user := range cfg.Users {
		if strings.TrimSpace(user.SSHKeyPath) == "" {
			continue
		}
		if err := validateKeyAlias(user.Alias); err != nil {
			return nil, err
		}

		privateKey, err := os.ReadFile(user.SSHKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key for %s: %w", user.Alias, err)
		}
		publicKey, err := os.ReadFile(user.SSHKeyPath + ".pub")
		if err != nil {
			return nil, fmt.Errorf("failed to read public key for %s: %w", user.Alias, err)
		}

		files = append(files,
			archiveKeyFile{
				Alias: user.Alias,
				Path:  keyArchivePath(user.Alias, false),
				Data:  privateKey,
				Mode:  0600,
			},
			archiveKeyFile{
				Alias: user.Alias,
				Path:  keyArchivePath(user.Alias, true),
				Data:  publicKey,
				Mode:  0644,
			},
		)
	}
	return files, nil
}

func keyArchivePath(alias string, public bool) string {
	if public {
		return filepath.ToSlash(filepath.Join(PayloadKeysDir, alias+".pub"))
	}
	return filepath.ToSlash(filepath.Join(PayloadKeysDir, alias))
}

func validateKeyAlias(alias string) error {
	if alias == "" || strings.Contains(alias, "/") || strings.Contains(alias, "\\") || alias == "." || alias == ".." || strings.Contains(alias, "..") {
		return fmt.Errorf("cannot export SSH key for unsafe alias %q", alias)
	}
	return nil
}
