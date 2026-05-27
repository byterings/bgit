package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/byterings/bgit/internal/platform"
	gossh "golang.org/x/crypto/ssh"
)

// GenerateSSHKey generates a new Ed25519 SSH key pair.
func GenerateSSHKey(username string) (privateKeyPath, publicKeyPath string, err error) {
	sshDir, err := platform.GetSSHDir()
	if err != nil {
		return "", "", err
	}

	if err := platform.MkdirSecure(sshDir); err != nil {
		return "", "", fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	privateKeyPath = filepath.Join(sshDir, fmt.Sprintf("bgit_%s", username))
	publicKeyPath = privateKeyPath + ".pub"

	if _, err := os.Stat(privateKeyPath); err == nil {
		return "", "", fmt.Errorf("key already exists at %s", privateKeyPath)
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	sshPubKey, err := gossh.NewPublicKey(pubKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to convert public key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: edPrivateKeyToPEM(privKey),
	}

	privateKeyFile, err := platform.OpenFileSecure(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return "", "", fmt.Errorf("failed to create private key file: %w", err)
	}
	defer privateKeyFile.Close()

	if err := pem.Encode(privateKeyFile, pemBlock); err != nil {
		return "", "", fmt.Errorf("failed to write private key: %w", err)
	}

	publicKeyBytes := gossh.MarshalAuthorizedKey(sshPubKey)
	if err := os.WriteFile(publicKeyPath, publicKeyBytes, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write public key: %w", err)
	}

	return privateKeyPath, publicKeyPath, nil
}

// ValidateSSHKeyPath checks if an SSH key exists and is readable.
func ValidateSSHKeyPath(path string) error {
	expandedPath, err := platform.ExpandTilde(path)
	if err != nil {
		return err
	}
	path = expandedPath

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("key file does not exist: %s", path)
		}
		return fmt.Errorf("failed to access key file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	ok, err := platform.CheckFilePermissions(path)
	if err != nil {
		return err
	}
	if !ok {
		mode := info.Mode()
		fmt.Printf("Warning: Key file has insecure permissions: %s\n", mode)
		fmt.Printf("  Run: %s\n", platform.GetPermissionFixCommand(path))
	}

	return nil
}

// GenerateSSHKeySystem uses system ssh-keygen for reliable key generation.
// Falls back to GenerateSSHKey if ssh-keygen is not available.
func GenerateSSHKeySystem(username string) (privateKeyPath, publicKeyPath string, err error) {
	if !platform.HasCommand("ssh-keygen") {
		fmt.Println("ssh-keygen not found, using built-in key generation...")
		return GenerateSSHKey(username)
	}

	sshDir, err := platform.GetSSHDir()
	if err != nil {
		return "", "", err
	}

	if err := platform.MkdirSecure(sshDir); err != nil {
		return "", "", fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	privateKeyPath = filepath.Join(sshDir, fmt.Sprintf("bgit_%s", username))
	publicKeyPath = privateKeyPath + ".pub"

	if _, err := os.Stat(privateKeyPath); err == nil {
		return "", "", fmt.Errorf("key already exists at %s", privateKeyPath)
	}

	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "", "-C", username+"@bgit")
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to generate SSH key: %w", err)
	}

	return privateKeyPath, publicKeyPath, nil
}

// GetPublicKeyContent reads and returns the public key content.
func GetPublicKeyContent(privateKeyPath string) (string, error) {
	publicKeyPath := privateKeyPath + ".pub"
	content, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %w", err)
	}
	return string(content), nil
}

func edPrivateKeyToPEM(key ed25519.PrivateKey) []byte {
	return []byte(key)
}
