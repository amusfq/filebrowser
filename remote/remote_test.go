package remote

import (
	"strings"
	"testing"
)

func TestCredentials(t *testing.T) {
	key := []byte("instance-key")
	server := Server{ID: "one"}
	if err := server.SetSecret("secret-password", key); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(server.Secret, "secret-password") {
		t.Fatal("credential was not encrypted")
	}
	if plain, err := server.secret(key); err != nil || plain != "secret-password" {
		t.Fatalf("decrypt: %q, %v", plain, err)
	}
	if _, err := server.secret([]byte("wrong-key")); err == nil {
		t.Fatal("accepted wrong key")
	}
	server.ID = "two"
	if _, err := server.secret(key); err == nil {
		t.Fatal("accepted credential copied from another connection")
	}
	if err := server.SetSecret("", key); err == nil {
		t.Fatal("accepted empty credential")
	}
	server.KeyAuth = true
	if err := server.SetSecret("invalid key", key); err == nil {
		t.Fatal("accepted invalid private key")
	}
}

func TestValidate(t *testing.T) {
	valid := Server{Name: "Storage", Host: "example.com", Username: "files", Root: "/srv/files/", Fingerprint: "SHA256:" + strings.Repeat("A", 43)}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	if valid.Port != 22 || valid.Root != "/srv/files" {
		t.Fatalf("defaults: %+v", valid)
	}
	for _, change := range []func(*Server){
		func(s *Server) { s.Root = "relative" },
		func(s *Server) { s.Root = "/srv\x00" },
		func(s *Server) { s.Fingerprint = "anything" },
		func(s *Server) { s.Port = 65536 },
		func(s *Server) { s.Host = "" },
	} {
		server := valid
		change(&server)
		if err := server.Validate(); err == nil {
			t.Fatalf("accepted invalid server: %+v", server)
		}
	}
}

func TestWithin(t *testing.T) {
	for _, test := range []struct {
		root, target string
		want         bool
	}{
		{"/srv/files", "/srv/files", true},
		{"/srv/files", "/srv/files/a", true},
		{"/srv/files", "/srv/files-other/secret", false},
		{"/srv/files", "/srv/secret", false},
		{"/", "/anything", true},
	} {
		if got := within(test.root, test.target); got != test.want {
			t.Errorf("within(%q, %q) = %v", test.root, test.target, got)
		}
	}
}
