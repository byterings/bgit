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

// ExportEnvelopeHeader describes the encrypted outer file wrapper.
type ExportEnvelopeHeader struct {
	FormatVersion string                  `json:"format_version"`
	BgitVersion   string                  `json:"bgit_version"`
	CreatedAt     string                  `json:"created_at"`
	Payload       ExportPayloadDescriptor `json:"payload"`
	Encryption    ExportEncryptionHeader  `json:"encryption"`
}

// ExportPayloadDescriptor identifies the preserved inner archive contract.
type ExportPayloadDescriptor struct {
	LayoutVersion string `json:"layout_version"`
	Compression   string `json:"compression"`
}

// ExportEncryptionHeader stores decryption metadata outside the ciphertext.
type ExportEncryptionHeader struct {
	KDF    ExportKDFHeader    `json:"kdf"`
	Cipher ExportCipherHeader `json:"cipher"`
}

type ExportKDFHeader struct {
	Algorithm   string `json:"algorithm"`
	SaltHex     string `json:"salt_hex"`
	TimeCost    uint32 `json:"time_cost"`
	MemoryKiB   uint32 `json:"memory_kib"`
	Parallelism uint8  `json:"parallelism"`
	KeyLength   uint32 `json:"key_length"`
}

type ExportCipherHeader struct {
	Algorithm string `json:"algorithm"`
	NonceHex  string `json:"nonce_hex"`
}

// ImportArchiveResult describes the config restored from an encrypted archive.
type ImportArchiveResult struct {
	UsersCount      int
	WorkspacesCount int
	BindingsCount   int
	ActiveUser      string
}
