// Package remote connects to SSH file storage using SFTP.
package remote

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Root        string `json:"root"`
	Fingerprint string `json:"fingerprint"`
	Secret      string `json:"secret,omitempty"`
	KeyAuth     bool   `json:"keyAuth"`
}

func (s *Server) Validate() error {
	if s.Port == 0 {
		s.Port = 22
	}
	s.Name = strings.TrimSpace(s.Name)
	s.Host = strings.TrimSpace(s.Host)
	if s.Name == "" || s.Host == "" || s.Username == "" || s.Port < 1 || s.Port > 65535 {
		return errors.New("name, host, username and a valid SSH port are required")
	}
	if !path.IsAbs(s.Root) || strings.ContainsRune(s.Root, '\x00') || strings.Contains(s.Root, "\\") {
		return errors.New("storage root must be an absolute POSIX path")
	}
	s.Root = path.Clean(s.Root)
	fingerprint, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(s.Fingerprint, "SHA256:"))
	if err != nil || len(fingerprint) != sha256.Size || !strings.HasPrefix(s.Fingerprint, "SHA256:") {
		return errors.New("a SHA256 SSH host-key fingerprint is required")
	}
	return nil
}

func secretCipher(key []byte) (cipher.AEAD, error) {
	if len(key) == 0 {
		return nil, errors.New("missing encryption key")
	}
	hash := sha256.Sum256(key)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (s *Server) SetSecret(secret string, key []byte) error {
	if secret == "" {
		return errors.New("password or private key is required")
	}
	if s.KeyAuth {
		if _, err := ssh.ParsePrivateKey([]byte(secret)); err != nil {
			return errors.New("invalid private key; use an unencrypted OpenSSH or PEM key")
		}
	}
	aead, err := secretCipher(key)
	if err != nil {
		return err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	s.Secret = base64.StdEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(secret), []byte(s.ID)))
	return nil
}

func (s Server) secret(key []byte) (string, error) {
	aead, err := secretCipher(key)
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(s.Secret)
	if err != nil || len(data) < aead.NonceSize() {
		return "", errors.New("invalid encrypted SSH credential")
	}
	plain, err := aead.Open(nil, data[:aead.NonceSize()], data[aead.NonceSize():], []byte(s.ID))
	return string(plain), err
}

func (s Server) Connect(ctx context.Context, key []byte) (*sftp.Client, func(), error) {
	secret, err := s.secret(key)
	if err != nil {
		return nil, nil, err
	}
	auth := ssh.Password(secret)
	if s.KeyAuth {
		signer, err := ssh.ParsePrivateKey([]byte(secret))
		if err != nil {
			return nil, nil, err
		}
		auth = ssh.PublicKeys(signer)
	}
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	cleanup := func() { stop(); _ = conn.Close() }
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, &ssh.ClientConfig{
		User: s.Username, Auth: []ssh.AuthMethod{auth},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			if ssh.FingerprintSHA256(key) != s.Fingerprint {
				return errors.New("SSH host-key fingerprint mismatch")
			}
			return nil
		},
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Minute))
	client := ssh.NewClient(sshConn, chans, reqs)
	fs, err := sftp.NewClient(client)
	if err != nil {
		cleanup()
		_ = client.Close()
		return nil, nil, err
	}
	return fs, func() { cleanup(); _ = fs.Close(); _ = client.Close() }, nil
}

// Resolve follows existing symlinks remotely before checking confinement.
// shortcut: SFTP checks are not atomic with open; use a server-side chroot when other processes can rename ancestors.
func Resolve(client *sftp.Client, root, name string) (string, error) {
	if strings.ContainsRune(name, '\x00') || strings.Contains(name, "\\") {
		return "", os.ErrPermission
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return "", os.ErrPermission
		}
	}
	base, err := client.RealPath(root)
	if err != nil {
		return "", err
	}
	base, err = resolveLinks(client, base)
	if err != nil {
		return "", err
	}
	target, err := resolveLinks(client, path.Join(base, strings.TrimPrefix(name, "/")))
	if err != nil {
		return "", err
	}
	if !within(base, target) {
		return "", fmt.Errorf("path escapes storage root: %w", os.ErrPermission)
	}
	return target, nil
}

func resolveLinks(client *sftp.Client, target string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(path.Clean(target), "/"), "/")
	resolved := "/"
	for hops := 0; len(parts) > 0; {
		next := path.Join(resolved, parts[0])
		parts = parts[1:]
		info, err := client.Lstat(next)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			resolved = next
			continue
		}
		hops++
		if hops > 40 {
			return "", os.ErrPermission
		}
		link, err := client.ReadLink(next)
		if err != nil {
			return "", err
		}
		if !path.IsAbs(link) {
			link = path.Join(resolved, link)
		}
		parts = append(strings.Split(strings.TrimPrefix(path.Clean(link), "/"), "/"), parts...)
		resolved = "/"
	}
	return resolved, nil
}

func within(root, target string) bool {
	return target == root || strings.HasPrefix(target, strings.TrimSuffix(root, "/")+"/")
}
