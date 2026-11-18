package fbhttp

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/remote"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/gorilla/mux"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func testSSHStorage(t *testing.T) (remote.Server, string) {
	t.Helper()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	config := &ssh.ServerConfig{PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		if meta.User() != "files" || string(password) != "test-password" {
			return nil, os.ErrPermission
		}
		return nil, nil
	}}
	config.AddHostKey(signer)
	config.PublicKeyCallback = func(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if meta.User() != "files" || string(key.Marshal()) != string(signer.PublicKey().Marshal()) {
			return nil, os.ErrPermission
		}
		return nil, nil
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				connection, channels, requests, err := ssh.NewServerConn(conn, config)
				if err != nil {
					return
				}
				defer connection.Close()
				go ssh.DiscardRequests(requests)
				for channel := range channels {
					stream, requests, err := channel.Accept()
					if err != nil {
						return
					}
					go func() {
						defer stream.Close()
						for request := range requests {
							if request.Type != "subsystem" {
								_ = request.Reply(false, nil)
								continue
							}
							_ = request.Reply(true, nil)
							server, err := sftp.NewServer(stream)
							if err != nil {
								return
							}
							defer server.Close()
							_ = server.Serve()
							return
						}
					}()
				}
			}()
		}
	}()
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	portNumber, _ := strconv.Atoi(port)
	privateBlock, err := ssh.MarshalPrivateKey(private, "test")
	if err != nil {
		t.Fatal(err)
	}
	return remote.Server{Name: "Test storage", Host: host, Port: portNumber, Username: "files", Root: t.TempDir(), Fingerprint: ssh.FingerprintSHA256(signer.PublicKey())}, string(pem.EncodeToMemory(privateBlock))
}

func TestRemoteStorage(t *testing.T) {
	sshServer, privateKey := testSSHStorage(t)
	if err := os.WriteFile(filepath.Join(sshServer.Root, "hello.txt"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(sshServer.Root, "folder"), 0700); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(sshServer.Root, "escape")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	key := []byte("test-key")
	perm := users.Permissions{Admin: true, Create: true, Download: true}
	store := scopedUserStorage(t, t.TempDir(), perm, key)
	token := signToken(t, perm, key)
	router := mux.NewRouter()
	router.Handle("/servers", handle(remoteServersHandler, "", store, &settings.Server{})).Methods("GET", "POST")
	router.Handle("/servers/{id}", handle(remoteServersHandler, "", store, &settings.Server{})).Methods("DELETE")
	router.Handle("/servers/{id}/files", handle(remoteFilesHandler, "", store, &settings.Server{})).Methods("GET", "POST")
	request := func(method, target, body, auth string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("X-Auth", auth)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	data, _ := json.Marshal(struct {
		remote.Server
		Credential string `json:"credential"`
	}{sshServer, "test-password"})
	if rec := request("POST", "/servers", string(data), ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated: %d", rec.Code)
	}
	rec := request("POST", "/servers", string(data), token)
	if rec.Code != http.StatusOK {
		t.Fatalf("save: %d %s", rec.Code, rec.Body.String())
	}
	var saved remote.Server
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" || saved.Secret != "" || strings.Contains(rec.Body.String(), "test-password") {
		t.Fatal("save response exposes credential or has no ID")
	}
	set, err := store.Settings.Get()
	if err != nil {
		t.Fatal(err)
	}
	if len(set.RemoteServers) != 1 || set.RemoteServers[0].Secret == "" || set.RemoteServers[0].Secret == "test-password" {
		t.Fatal("credential not encrypted in storage")
	}
	endpoint := "/servers/" + saved.ID + "/files"
	for _, test := range []struct {
		name, method, target, body string
		status                     int
	}{
		{"list", "GET", endpoint, "", 200},
		{"download", "GET", endpoint + "?download=true&path=/hello.txt", "", 200},
		{"upload", "POST", endpoint + "?path=/folder&name=new.txt", "new content", 201},
		{"overwrite", "POST", endpoint + "?path=/folder&name=new.txt", "overwritten", 409},
		{"invalid filename", "POST", endpoint + "?name=../outside.txt", "bad", 400},
		{"traversal", "GET", endpoint + "?path=" + url.QueryEscape("/../"), "", 403},
		{"symlink escape", "GET", endpoint + "?path=/escape/secret.txt&download=true", "", 403},
		{"symlink upload", "POST", endpoint + "?path=/escape&name=new.txt", "bad", 403},
		{"missing server", "GET", "/servers/missing/files", "", 404},
	} {
		t.Run(test.name, func(t *testing.T) {
			rec := request(test.method, test.target, test.body, token)
			if rec.Code != test.status {
				t.Fatalf("status %d, expected %d: %s", rec.Code, test.status, rec.Body.String())
			}
			if test.name == "list" && strings.Contains(rec.Body.String(), "escape") {
				t.Fatal("listing exposes escaping symlink")
			}
			if test.name == "download" && rec.Body.String() != "hello" {
				t.Fatalf("download: %q", rec.Body.String())
			}
		})
	}
	content, err := os.ReadFile(filepath.Join(sshServer.Root, "folder", "new.txt"))
	if err != nil || string(content) != "new content" {
		t.Fatalf("upload corrupted: %q %v", content, err)
	}
	list := request("GET", "/servers", "", token)
	if list.Code != 200 || strings.Contains(list.Body.String(), "secret") {
		t.Fatalf("unsafe server list: %s", list.Body.String())
	}
	user, err := store.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatal(err)
	}
	user.Perm.Create = false
	user.Perm.Download = false
	if err := store.Users.Update(user, "Perm"); err != nil {
		t.Fatal(err)
	}
	if rec := request("POST", endpoint+"?name=denied.txt", "content", token); rec.Code != 403 {
		t.Fatalf("create permission ignored: %d", rec.Code)
	}
	if rec := request("GET", endpoint+"?download=true&path=/hello.txt", "", token); rec.Code != 403 {
		t.Fatalf("download permission ignored: %d", rec.Code)
	}
	user.Perm.Admin = false
	if err := store.Users.Update(user, "Perm"); err != nil {
		t.Fatal(err)
	}
	if rec := request("GET", endpoint, "", token); rec.Code != 403 {
		t.Fatalf("non-admin access: %d", rec.Code)
	}
	user.Perm = perm
	if err := store.Users.Update(user, "Perm"); err != nil {
		t.Fatal(err)
	}
	set.RemoteServers[0].Fingerprint = "SHA256:" + strings.Repeat("A", 43)
	if err := store.Settings.Save(set); err != nil {
		t.Fatal(err)
	}
	if rec := request("GET", endpoint, "", token); rec.Code != 502 {
		t.Fatalf("accepted wrong host key: %d", rec.Code)
	}
	if rec := request("DELETE", "/servers/"+saved.ID, "", token); rec.Code != 200 {
		t.Fatalf("delete connection: %d", rec.Code)
	}
	if rec := request("GET", endpoint, "", token); rec.Code != 404 {
		t.Fatalf("connection still exists: %d", rec.Code)
	}
	if _, err := os.Stat(filepath.Join(sshServer.Root, "hello.txt")); err != nil {
		t.Fatal("removing connection removed files")
	}
	data, _ = json.Marshal(struct {
		remote.Server
		Credential string `json:"credential"`
	}{sshServer, "wrong-password"})
	if rec := request("POST", "/servers", string(data), token); rec.Code != 502 {
		t.Fatalf("accepted invalid SSH password: %d", rec.Code)
	}
	sshServer.KeyAuth = true
	data, _ = json.Marshal(struct {
		remote.Server
		Credential string `json:"credential"`
	}{sshServer, privateKey})
	rec = request("POST", "/servers", string(data), token)
	if rec.Code != 200 {
		t.Fatalf("private key authentication: %d %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if rec := request("GET", "/servers/"+saved.ID+"/files", "", token); rec.Code != 200 {
		t.Fatalf("saved private key connection failed: %d", rec.Code)
	}
}
