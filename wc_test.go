//
// nazuna :: wc_test.go
//
//   Copyright (c) 2013-2018 Akinori Hattori <hattya@gmail.com>
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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestWC(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	if err := mkdir(".nzn", "r", ".git"); err != nil {
		t.Fatal(err)
	}
	repo, err := nazuna.Open(nil, ".")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(".nzn", "state.json")

	if err := mkdir(path); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.WC(); err == nil {
		t.Error("expected error")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	wc, err := repo.WC()
	if err != nil {
		t.Fatal(err)
	}
	if err := wc.Flush(); err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if g, e := string(data), "{}\n"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func TestWCError(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	if err := mkdir(".nzn", "r", ".git"); err != nil {
		t.Fatal(err)
	}
	repo, err := nazuna.Open(nil, ".")
	if err != nil {
		t.Fatal(err)
	}

	wc, err := repo.WC()
	if err != nil {
		t.Fatal(err)
	}
	switch _, err := wc.LayerFor("_"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasSuffix(err.Error(), "layer '_'"):
		t.Error("unexpected error:", err)
	}
	switch err := wc.Unlink("_"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasSuffix(err.Error(), ": path is not link"):
		t.Error("unexpected error:", err)
	}
}

var entryTests = []struct {
	e *nazuna.Entry
	s string
}{
	{
		&nazuna.Entry{},
		"!",
	},
	{
		&nazuna.Entry{
			IsDir: true,
		},
		"!",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
		},
		"!layer",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
			IsDir: true,
		},
		"!layer",
	},
	{
		&nazuna.Entry{
			Path: "path",
		},
		"path!",
	},
	{
		&nazuna.Entry{
			Path:  "path",
			IsDir: true,
		},
		"path/!",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
			Path:  "path",
		},
		"path!layer",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
			Path:  "path",
			IsDir: true,
		},
		"path/!layer",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
		},
		"path!layer:origin",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
			IsDir:  true,
		},
		"path/!layer:origin/",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
			Type:   "link",
		},
		"path!origin",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
			IsDir:  true,
			Type:   "link",
		},
		"path/!origin" + string(os.PathSeparator),
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "github.com/hattya/nazuna",
			Type:   "subrepo",
		},
		"path!github.com/hattya/nazuna",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "github.com/hattya/nazuna",
			IsDir:  true,
			Type:   "subrepo",
		},
		"path/!github.com/hattya/nazuna",
	},
}

func TestEntry(t *testing.T) {
	for _, tt := range entryTests {
		if g, e := tt.e.Format("%v!%v"), tt.s; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
}
