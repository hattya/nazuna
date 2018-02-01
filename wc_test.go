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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestOpenWC(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())
	popd, err := pushd(repo.Root())
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	wc, err := repo.WC()
	if err != nil {
		t.Fatal(err)
	}
	if err := wc.Flush(); err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile(filepath.Join(".nzn", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if g, e := string(data), "{}\n"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

func TestOpenWCError(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())
	popd, err := pushd(repo.Root())
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	// unmarshal error
	if err := mkdir(".nzn", "state.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.WC(); err == nil {
		t.Error("expected error")
	}
}

func TestWCPaths(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())
	popd, err := pushd(repo.Root())
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	wc, err := repo.WC()
	if err != nil {
		t.Fatal(err)
	}
	if g, e := wc.PathFor("file"), filepath.Join(repo.Root(), "file"); g != e {
		t.Errorf("WC.PathFor(%q) = %q, expected %q", "file", g, e)
	}
	if wc.Exists("file") {
		t.Errorf("WC.Exists(%q) = true, expected false", "file")
	}
	// base: /
	base := '/'
	if rel, err := wc.Rel(base, filepath.Join(repo.Root(), "file")); err != nil {
		t.Error(err)
	} else if g, e := rel, "file"; g != e {
		t.Errorf("WC.Rel('%q', ...) = %q, expected %q", base, g, e)
	}
	if rel, err := wc.Rel(base, "file"); err != nil {
		t.Error(err)
	} else if g, e := rel, "file"; g != e {
		t.Errorf("WC.Rel('%q', ...) = %q, expected %q", base, g, e)
	}
	// base: .
	base = '.'
	if rel, err := wc.Rel(base, "file"); err != nil {
		t.Error(err)
	} else if g, e := rel, "file"; g != e {
		t.Errorf("WC.Rel('%q', ...) = %q, expected %q", base, g, e)
	}
	if rel, err := wc.Rel(base, "$var"); err != nil {
		t.Error(err)
	} else if g, e := rel, "$var"; g != e {
		t.Errorf("WC.Rel('%q', ...) = %q, expected %q", base, g, e)
	}
	// unknown base
	if _, err := wc.Rel('_', ""); err == nil {
		t.Error("expected error")
	}
	// not under root
	if _, err := wc.Rel('/', filepath.Dir(repo.Root())); err == nil {
		t.Error("expected error")
	}
	if _, err := wc.Rel('/', filepath.Join(filepath.Dir(repo.Root()), "file")); err == nil {
		t.Error("expected error")
	}
}

func TestWCLinks(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())
	popd, err := pushd(repo.Root())
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	wc, err := repo.WC()
	if err != nil {
		t.Fatal(err)
	}
	// file
	dst := "link"
	src := repo.PathFor(nil, dst)
	if err := touch(src); err != nil {
		t.Fatal(err)
	}
	if err := wc.Link(src, dst); err != nil {
		t.Fatal(err)
	}
	if err := testLink(wc, src, dst); err != nil {
		t.Error(err)
	}
	if err := wc.Unlink(dst); err != nil {
		t.Fatal(err)
	}
	// file in directory
	dst = filepath.Join("dir", "link")
	src = repo.PathFor(nil, dst)
	if err := mkdir(filepath.Dir(src)); err != nil {
		t.Fatal(err)
	}
	if err := touch(src); err != nil {
		t.Fatal(err)
	}
	if err := wc.Link(src, dst); err != nil {
		t.Fatal(err)
	}
	if err := testLink(wc, src, dst); err != nil {
		t.Error(err)
	}
	if err := wc.Unlink(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(dst)); err == nil {
		t.Fatal("expected to remove parent directories")
	}
	// directory
	dst = "dir"
	src = repo.PathFor(nil, dst)
	if err := wc.Link(src, dst); err != nil {
		t.Fatal(err)
	}
	if err := testLink(wc, src, dst); err != nil {
		t.Error(err)
	}
	if err := wc.Unlink(dst); err != nil {
		t.Fatal(err)
	}
	// keep non-empty directory
	dst = filepath.Join("dir", "link")
	src = repo.PathFor(nil, dst)
	if err := wc.Link(src, dst); err != nil {
		t.Fatal(err)
	}
	if err := touch(filepath.Join("dir", "file")); err != nil {
		t.Fatal(err)
	}
	if err := wc.Unlink(dst); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat("dir"); err != nil {
		t.Error("expected to keep parent directories")
	}
	if err := os.RemoveAll("dir"); err != nil {
		t.Fatal(err)
	}
	// parent path is link
	dst = "dir"
	src = repo.PathFor(nil, dst)
	if err := wc.Link(src, dst); err != nil {
		t.Fatal(err)
	}
	dst = filepath.Join("dir", "file")
	src = repo.PathFor(nil, dst)
	switch err := wc.Link(src, dst).(type) {
	case *os.PathError:
		if g, e := err.Err, nazuna.ErrLink; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	default:
		t.Errorf("expected *os.PathError, got %T", err)
	}
	if err := wc.Unlink(filepath.Join("dir", "file")); err == nil {
		t.Error("expected error")
	}
	if err := wc.Unlink("dir"); err != nil {
		t.Fatal(err)
	}
}

func testLink(wc *nazuna.WC, src, dst string) error {
	if !wc.IsLink(dst) {
		return fmt.Errorf("wc.IsLink(%q) = false, expected true", dst)
	}
	if !wc.LinksTo(dst, src) {
		return fmt.Errorf("wc.LinksTo(%q) = false, expected true", dst)
	}
	if err := wc.Link(src, dst); err == nil {
		return fmt.Errorf("expected error")
	}
	return nil
}

func TestSelectLayer(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())
	popd, err := pushd(repo.Root())
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	wc, err := repo.WC()
	if err != nil {
		t.Fatal(err)
	}
	master, err := repo.NewLayer("master")
	if err != nil {
		t.Fatal(err)
	}
	linux, err := repo.NewLayer("os/linux")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.NewLayer("os/windows"); err != nil {
		t.Fatal(err)
	}
	// cannot resolve
	switch _, err := wc.LayerFor("os"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "cannot resolve layer "):
		t.Error("unexpected error:", err)
	}
	switch _, err := wc.Layers(); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "cannot resolve layer "):
		t.Error("unexpected error:", err)
	}
	// cannot select
	for _, s := range []string{"_", "master", "os"} {
		if err := wc.SelectLayer(s); err == nil {
			t.Errorf("%v: expected error", s)
		}
	}

	if err := wc.SelectLayer("os/windows"); err != nil {
		t.Error(err)
	}
	if err := wc.SelectLayer("os/linux"); err != nil {
		t.Error(err)
	}
	if _, err := wc.LayerFor("os"); err != nil {
		t.Error(err)
	}
	if ll, err := wc.Layers(); err != nil {
		t.Error(err)
	} else if g, e := ll, []*nazuna.Layer{linux, master}; !reflect.DeepEqual(g, e) {
		t.Errorf("WC.Layers() = {%q, %q}, expected {%q, %q}", g[0].Path(), g[1].Path(), e[0].Path(), e[1].Path())
	}
	// already selected
	if err := wc.SelectLayer("os/linux"); err == nil {
		t.Error("expected error")
	}
}

func TestWCErrorf(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	wc, err := repo.WC()
	if err != nil {
		t.Fatal(err)
	}

	err = fmt.Errorf("error")
	if g, e := wc.Errorf(err).Error(), "error"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	err = &os.LinkError{
		New: filepath.Join(repo.Root(), "link"),
		Err: fmt.Errorf("link error"),
	}
	if g, e := wc.Errorf(err).Error(), "link: link error"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	err = &os.PathError{
		Path: filepath.Join(repo.Root(), "link"),
		Err:  fmt.Errorf("path error"),
	}
	if g, e := wc.Errorf(err).Error(), "link: path error"; g != e {
		t.Errorf("expected %q, got %q", e, g)
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
