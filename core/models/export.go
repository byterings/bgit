package models

// ExportManifest describes the top-level metadata stored in a .bgit archive.
type ExportManifest struct {
	FormatVersion string                  `json:"format_version"`
	BgitVersion   string                  `json:"bgit_version"`
	CreatedAt     string                  `json:"created_at"`
	Archive       ExportArchiveDescriptor `json:"archive"`
	Identities    []ExportIdentitySummary `json:"identities"`
}

// ExportArchiveDescriptor describes the stable archive layout used by exports.
type ExportArchiveDescriptor struct {
	LayoutVersion string                `json:"layout_version"`
	Encryption    ExportEncryptionState `json:"encryption"`
	Files         []ExportArchiveFile   `json:"files"`
}

// ExportEncryptionState documents the current and planned protection state.
type ExportEncryptionState struct {
	Status         string `json:"status"`
	PlannedVersion string `json:"planned_version"`
}

// ExportArchiveFile describes a file or reserved directory inside the archive.
type ExportArchiveFile struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Required bool   `json:"required"`
}

// ExportIdentitySummary captures the identities represented in the archive.
type ExportIdentitySummary struct {
	Alias          string `json:"alias"`
	GitHubUsername string `json:"github_username"`
	HasSSHKeyPath  bool   `json:"has_ssh_key_path"`
}
