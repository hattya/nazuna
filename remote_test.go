//
// nazuna :: remote_test.go
//
//   Copyright (c) 2013 Akinori Hattori <hattya@gmail.com>
//
//   Permission is hereby granted, free of charge, to any person
//   obtaining a copy of this software and associated documentation files
//   (the "Software"), to deal in the Software without restriction,
//   including without limitation the rights to use, copy, modify, merge,
//   publish, distribute, sublicense, and/or sell copies of the Software,
//   and to permit persons to whom the Software is furnished to do so,
//   subject to the following conditions:
//
//   The above copyright notice and this permission notice shall be
//   included in all copies or substantial portions of the Software.
//
//   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//   EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
//   MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//   NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
//   BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
//   ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
//   CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//   SOFTWARE.
//

package nazuna_test

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func newHTTPClient(addr string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(network, _ string) (net.Conn, error) {
				return net.Dial(network, addr)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func TestRemote(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/1.0/repositories/") {
			l := strings.Split(r.RequestURI[18:], "/")
			fmt.Fprintf(w, `{"owner":"%s","scm":"%s"}`, l[0], l[1])
		}
	}))
	defer s.Close()

	c := http.DefaultClient
	defer func() { http.DefaultClient = c }()
	http.DefaultClient = newHTTPClient(s.Listener.Addr().String())

	r, err := nazuna.NewRemote("github.com/kien/ctrlp.vim")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := r.VCS.Cmd, "git"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := r.URI, "https://github.com/kien/ctrlp.vim"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := r.Path, ""; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	r, err = nazuna.NewRemote("bitbucket.org/hattya/git")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := r.VCS.Cmd, "git"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := r.URI, "https://bitbucket.org/hattya/git.git"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := r.Path, ""; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	r, err = nazuna.NewRemote("bitbucket.org/hattya/hg")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := r.VCS.Cmd, "hg"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := r.URI, "https://bitbucket.org/hattya/hg"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := r.Path, ""; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

func TestRemoteError(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/1.0/repositories/") {
			l := strings.Split(r.RequestURI[18:], "/")
			if l[0] != "_" {
				switch l[1] {
				case "_":
					http.NotFound(w, r)
				default:
					fmt.Fprintf(w, `{"owner":"%s","scm":"%s"}`, l[0], l[1])
				}
			}
		}
	}))
	defer s.Close()

	c := http.DefaultClient
	defer func() { http.DefaultClient = c }()
	http.DefaultClient = newHTTPClient(s.Listener.Addr().String())

	switch _, err := nazuna.NewRemote("github.com/hattya"); {
	case err == nil:
		t.Error("expected error")
	case err != nazuna.ErrRemote:
		t.Error("unexpected error:", err)
	}
	switch _, err := nazuna.NewRemote("bitbucket.org/hattya/svn"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "cannot detect remote vcs "):
		t.Error("unexpected error:", err)
	}
	switch _, err := nazuna.NewRemote("bitbucket.org/hattya/_"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasSuffix(err.Error(), "/hattya/_: 404 Not Found"):
		t.Error("unexpected error:", err)
	}
	switch _, err := nazuna.NewRemote("bitbucket.org/_/_"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasSuffix(err.Error(), "/_/_: unexpected end of JSON input"):
		t.Error("unexpected error:", err)
	}
}
