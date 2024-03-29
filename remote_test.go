//
// nazuna :: remote_test.go
//
//   Copyright (c) 2013-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
	"github.com/hattya/nazuna/internal/test"
)

var remoteTests = []struct {
	src                  string
	vcs, uri, root, path string
}{
	{
		src:  "github.com/mattn/gist-vim",
		vcs:  "git",
		uri:  "https://github.com/mattn/gist-vim",
		root: "github.com/mattn/gist-vim",
	},
	{
		src:  "bitbucket.org/hattya/git",
		vcs:  "git",
		uri:  "https://bitbucket.org/hattya/git.git",
		root: "bitbucket.org/hattya/git",
	},
	{
		src:  "bitbucket.org/hattya/hg",
		vcs:  "hg",
		uri:  "https://bitbucket.org/hattya/hg",
		root: "bitbucket.org/hattya/hg",
	},
}

func TestNewRemote(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/2.0/repositories/") {
			l := strings.Split(r.RequestURI[18:], "/")
			fmt.Fprintf(w, `{"owner":{"username":"%v"},"scm":"%v"}`, l[0], l[1])
		}
	}))
	defer s.Close()

	save := http.DefaultClient
	defer func() { http.DefaultClient = save }()
	http.DefaultClient = test.NewHTTPClient(s.Listener.Addr().String())

	for _, tt := range remoteTests {
		r, err := nazuna.NewRemote(nil, tt.src)
		if err != nil {
			t.Fatal(err)
		}
		if g, e := r.VCS, tt.vcs; g != e {
			t.Errorf("Remove.VCS = %v, expected %v", e, g)
		}
		if g, e := r.URI, tt.uri; g != e {
			t.Errorf("Remove.URI = %v, expected %v", e, g)
		}
		if g, e := r.Root, tt.root; g != e {
			t.Errorf("Remove.Root = %v, expected %v", e, g)
		}
		if g, e := r.Path, tt.path; g != e {
			t.Errorf("Remove.Path = %v, expected %v", e, g)
		}
	}
}

func TestNewRemoteError(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/2.0/repositories/") {
			l := strings.Split(r.RequestURI[18:], "/")
			if l[0] != "_" {
				if l[1] != "_" {
					fmt.Fprintf(w, `{"owner":{"username":"%v"},"scm":"%v"}`, l[0], l[1])
				} else {
					http.NotFound(w, r)
				}
			}
		}
	}))
	defer s.Close()

	save := http.DefaultClient
	defer func() { http.DefaultClient = save }()
	http.DefaultClient = test.NewHTTPClient(s.Listener.Addr().String())

	if _, err := nazuna.NewRemote(nil, "github.com/hattya"); err != nazuna.ErrRemote {
		t.Errorf("expected ErrRemote, got %v", err)
	}

	switch _, err := nazuna.NewRemote(nil, "bitbucket.org/hattya/_"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasSuffix(err.Error(), "/hattya/_: 404 Not Found"):
		t.Error("unexpected error:", err)
	}

	switch _, err := nazuna.NewRemote(nil, "bitbucket.org/_/_"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasSuffix(err.Error(), "/_/_: unexpected end of JSON input"):
		t.Error("unexpected error:", err)
	}
}

func TestRemote(t *testing.T) {
	dir := sandbox(t)
	home := filepath.Join(dir, "home")
	if err := mkdir(home); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	ts := httptest.NewTLSServer(http.FileServer(http.Dir(filepath.Join(dir, "public"))))
	defer ts.Close()

	git(t, "config", "--global", "user.name", "Nazuna")
	git(t, "config", "--global", "user.email", "nazuna@example.com")
	git(t, "config", "--global", "http.sslVerify", "false")
	git(t, "config", "--global", "url."+ts.URL+"/gist-vim/.git.insteadOf", "https://github.com/mattn/gist-vim")
	git(t, "init", "-q", filepath.Join(dir, "public", "gist-vim"))
	if err := os.Chdir(filepath.Join(dir, "public", "gist-vim")); err != nil {
		t.Fatal(err)
	}
	if err := touch("README.md"); err != nil {
		t.Fatal(err)
	}
	git(t, "add", ".")
	git(t, "commit", "-qm", ".")
	git(t, "update-server-info")
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	ui := new(testUI)
	r, err := nazuna.NewRemote(ui, "github.com/mattn/gist-vim")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Update(filepath.Base(r.Root)); err == nil {
		t.Error("expected error")
	}
	if err := r.Clone(home, filepath.Base(r.Root)); err != nil {
		t.Error(err)
	}
	if err := r.Update(filepath.Join(home, filepath.Base(r.Root))); err != nil {
		t.Log(ui.String())
		t.Error(err)
	}

	r.VCS = "cvs"
	if err := r.Clone(home, filepath.Base(r.Root)); err == nil {
		t.Error("expected error")
	}
}

func git(t *testing.T, a ...string) {
	t.Helper()

	var b bytes.Buffer
	cmd := exec.Command("git", a...)
	cmd.Stdout = &b
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		t.Log(b.String())
		t.Fatal(err)
	}
}
