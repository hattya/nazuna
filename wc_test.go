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
	a, err := repo.NewLayer("a")
	if err != nil {
		t.Fatal(err)
	}
	b1, err := repo.NewLayer("b/1")
	if err != nil {
		t.Fatal(err)
	}
	b2, err := repo.NewLayer("b/2")
	if err != nil {
		t.Fatal(err)
	}
	// cannot resolve
	switch _, err := wc.LayerFor("b"); {
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
	for _, s := range []string{"_", "a", "b"} {
		if err := wc.SelectLayer(s); err == nil {
			t.Errorf("%v: expected error", s)
		}
	}

	if err := wc.SelectLayer(b2.Path()); err != nil {
		t.Error(err)
	}
	if err := wc.SelectLayer(b1.Path()); err != nil {
		t.Error(err)
	}
	if _, err := wc.LayerFor("b"); err != nil {
		t.Error(err)
	}
	if ll, err := wc.Layers(); err != nil {
		t.Error(err)
	} else if g, e := ll, []*nazuna.Layer{b1, a}; !reflect.DeepEqual(g, e) {
		t.Errorf("WC.Layers() = {%q, %q}, expected {%q, %q}", g[0].Path(), g[1].Path(), e[0].Path(), e[1].Path())
	}
	// already selected
	if err := wc.SelectLayer(b1.Path()); err == nil {
		t.Error("expected error")
	}
}

func TestMergeLayers(t *testing.T) {
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
	a, err := repo.NewLayer("a")
	if err != nil {
		t.Fatal(err)
	}
	b1, err := repo.NewLayer("b/1")
	if err != nil {
		t.Fatal(err)
	}
	b2, err := repo.NewLayer("b/2")
	if err != nil {
		t.Fatal(err)
	}

	ui := new(testUI)
	git, err := nazuna.VCSFor(ui, filepath.Join(".nzn", "r"))
	if err != nil {
		t.Fatal(err)
	}
	alias := func(l *nazuna.Layer, src, dst string) {
		t.Helper()
		if err := l.NewAlias(src, dst); err != nil {
			t.Fatal(err)
		}
	}
	file := func(l *nazuna.Layer, n string) {
		t.Helper()
		if err := mkdir(filepath.Dir(repo.PathFor(l, n))); err != nil {
			t.Fatal(err)
		}
		if err := touch(repo.PathFor(l, n)); err != nil {
			t.Fatal(err)
		}
	}
	link := func(l *nazuna.Layer, src, dst string) {
		t.Helper()
		if _, err := l.NewLink(nil, repo.PathFor(l, src), dst+"a"); err != nil {
			t.Fatal(err)
		}
		if _, err := l.NewLink([]string{repo.PathFor(l, ".")}, src, dst+"b"); err != nil {
			t.Fatal(err)
		}
	}
	subrepo := func(l *nazuna.Layer, dst string) {
		t.Helper()
		src := "github.com/hattya/" + filepath.Base(dst)
		if _, err := l.NewSubrepo(src+"a", dst+"a"); err != nil {
			t.Fatal(err)
		}
		repo, err := l.NewSubrepo(src+"b", dst+"b")
		if err != nil {
			t.Fatal(err)
		}
		repo.Name = filepath.Base(repo.Src)
	}
	// file
	file(a, "file1")
	// file :: → alias
	file(a, "file2")
	alias(b1, "file2", "file2_")
	alias(b2, "file2", "file2_")
	// :: file
	file(b1, "file3")
	file(b2, "file3")
	// dir
	file(a, filepath.Join("dir1", "file1"))
	// dir/file
	file(a, filepath.Join("dir2", "file1"))
	// file :: → alias
	file(a, "file4")
	alias(b1, "file4", filepath.Join("dir2", "file2"))
	alias(b2, "file4", filepath.Join("dir2", "file2"))
	// dir/dir/file
	file(a, filepath.Join("dir2", "dir1", "file1"))
	// file :: → alias
	file(a, "file5")
	alias(b1, "file5", filepath.Join("dir2", "dir1", "file2"))
	alias(b2, "file5", filepath.Join("dir2", "dir1", "file2"))
	// dir/[file :: → alias]
	file(a, filepath.Join("dir2", "file3"))
	alias(b1, filepath.Join("dir2", "file3"), filepath.Join("dir2", "file3_"))
	alias(b2, filepath.Join("dir2", "file3"), filepath.Join("dir2", "file3_"))
	// dir/file :: → alias
	file(a, filepath.Join("dir3", "file1"))
	alias(b1, filepath.Join("dir3", "file1"), filepath.Join("dir2", "file4"))
	alias(b2, filepath.Join("dir3", "file1"), filepath.Join("dir2", "file4"))
	// [dir :: → alias]/file
	file(a, filepath.Join("dir4", "file5"))
	alias(b1, "dir4", "dir2")
	alias(b2, "dir4", "dir2")
	// :: dir/file
	file(b1, filepath.Join("dir2", "file6"))
	file(b2, filepath.Join("dir2", "file6"))
	// dir/file :: → alias
	file(a, filepath.Join("dir5", "file1"))
	alias(b1, filepath.Join("dir5", "file1"), "file6")
	alias(b2, filepath.Join("dir5", "file1"), "file6")
	// dir exists
	if err := mkdir(wc.PathFor("dir6")); err != nil {
		t.Fatal(err)
	}
	file(a, filepath.Join("dir6", "file1"))
	// link
	link(a, "file1", "link1")
	// link :: → alias
	link(a, "file2", "link2")
	alias(b1, "link2a", "link2a_")
	alias(b1, "link2b", "link2b_")
	alias(b2, "link2a", "link2a_")
	alias(b2, "link2b", "link2b_")
	// :: link
	link(b1, "file3", "link3")
	link(b2, "file3", "link3")
	// link :: → alias
	link(a, "file4", "link4")
	alias(b1, "link4a", filepath.Join("dir2", "link2a"))
	alias(b1, "link4b", filepath.Join("dir2", "link2b"))
	alias(b2, "link4a", filepath.Join("dir2", "link2a"))
	alias(b2, "link4b", filepath.Join("dir2", "link2b"))
	// dir/[link :: → alias]
	link(a, filepath.Join("dir2", "file3"), filepath.Join("dir2", "link3"))
	alias(b1, filepath.Join("dir2", "link3a"), filepath.Join("dir2", "link3a_"))
	alias(b1, filepath.Join("dir2", "link3b"), filepath.Join("dir2", "link3b_"))
	alias(b2, filepath.Join("dir2", "link3a"), filepath.Join("dir2", "link3a_"))
	alias(b2, filepath.Join("dir2", "link3b"), filepath.Join("dir2", "link3b_"))
	// dir/link :: → alias
	link(a, filepath.Join("dir3", "file1"), filepath.Join("dir3", "link1"))
	alias(b1, filepath.Join("dir3", "link1a"), filepath.Join("dir2", "link4a"))
	alias(b1, filepath.Join("dir3", "link1b"), filepath.Join("dir2", "link4b"))
	alias(b2, filepath.Join("dir3", "link1a"), filepath.Join("dir2", "link4a"))
	alias(b2, filepath.Join("dir3", "link1b"), filepath.Join("dir2", "link4b"))
	// [dir :: → alias]/link
	link(a, filepath.Join("dir4", "file5"), filepath.Join("dir4", "link5"))
	// :: dir/link
	link(b1, filepath.Join("dir2", "file6"), filepath.Join("dir2", "link6"))
	link(b2, filepath.Join("dir2", "file6"), filepath.Join("dir2", "link6"))
	// dir/link :: → alias
	link(a, filepath.Join("dir5", "file1"), filepath.Join("dir5", "link1"))
	alias(b1, filepath.Join("dir5", "link1a"), "link6a")
	alias(b1, filepath.Join("dir5", "link1b"), "link6b")
	alias(b2, filepath.Join("dir5", "link1a"), "link6a")
	alias(b2, filepath.Join("dir5", "link1b"), "link6b")
	// subrepo
	subrepo(a, "repo1")
	// subrepo :: → alias
	subrepo(a, "repo2")
	alias(b1, "repo2a", "repo2a_")
	alias(b1, "repo2b", "repo2b_")
	alias(b2, "repo2a", "repo2a_")
	alias(b2, "repo2b", "repo2b_")
	// :: subrepo
	subrepo(b1, "repo3")
	subrepo(b2, "repo3")
	// subrepo :: → alias
	subrepo(a, "repo4")
	alias(b1, "repo4a", filepath.Join("dir2", "repo1a"))
	alias(b1, "repo4b", filepath.Join("dir2", "repo1b"))
	alias(b2, "repo4a", filepath.Join("dir2", "repo1a"))
	alias(b2, "repo4b", filepath.Join("dir2", "repo1b"))
	// dir[subrepo :: → alias]
	subrepo(a, filepath.Join("dir2", "repo3"))
	alias(b1, filepath.Join("dir2", "repo3a"), filepath.Join("dir2", "repo3a_"))
	alias(b1, filepath.Join("dir2", "repo3b"), filepath.Join("dir2", "repo3b_"))
	alias(b2, filepath.Join("dir2", "repo3a"), filepath.Join("dir2", "repo3a_"))
	alias(b2, filepath.Join("dir2", "repo3b"), filepath.Join("dir2", "repo3b_"))
	// dir/subrepo → alias
	subrepo(a, filepath.Join("dir3", "repo1"))
	alias(b1, filepath.Join("dir3", "repo1a"), filepath.Join("dir2", "repo4a"))
	alias(b1, filepath.Join("dir3", "repo1b"), filepath.Join("dir2", "repo4b"))
	alias(b2, filepath.Join("dir3", "repo1a"), filepath.Join("dir2", "repo4a"))
	alias(b2, filepath.Join("dir3", "repo1b"), filepath.Join("dir2", "repo4b"))
	// [dir :: → alias]/subrepo
	subrepo(a, filepath.Join("dir4", "repo5"))
	// :: dir/subrepo
	subrepo(b1, filepath.Join("dir2", "repo6"))
	subrepo(b2, filepath.Join("dir2", "repo6"))
	// dir/subrepo :: → alias
	subrepo(a, filepath.Join("dir5", "repo1"))
	alias(b1, filepath.Join("dir5", "repo1a"), "repo6a")
	alias(b1, filepath.Join("dir5", "repo1b"), "repo6b")
	alias(b2, filepath.Join("dir5", "repo1a"), "repo6a")
	alias(b2, filepath.Join("dir5", "repo1b"), "repo6b")

	if err := git.Add("."); err != nil {
		t.Fatal(err)
	}

	wc.State.WC = []*nazuna.Entry{
		{
			Layer: a.Path(),
			Path:  "dir2",
			IsDir: true,
		},
	}
	e := []*nazuna.Entry{
		{
			Layer: a.Path(),
			Path:  "dir1",
			IsDir: true,
		},
		{
			Layer: a.Path(),
			Path:  "dir2/dir1/file1",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/dir1/file2",
			Origin: "file5",
		},
		{
			Layer: a.Path(),
			Path:  "dir2/file1",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/file2",
			Origin: "file4",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/file3_",
			Origin: "dir2/file3",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/file4",
			Origin: "dir3/file1",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/file5",
			Origin: "dir4/file5",
		},
		{
			Layer: b1.Path(),
			Path:  "dir2/file6",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link2a",
			Origin: repo.PathFor(a, "file4"),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link2b",
			Origin: repo.PathFor(a, "file4"),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link3a_",
			Origin: repo.PathFor(a, filepath.Join("dir2", "file3")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link3b_",
			Origin: repo.PathFor(a, filepath.Join("dir2", "file3")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link4a",
			Origin: repo.PathFor(a, filepath.Join("dir3", "file1")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link4b",
			Origin: repo.PathFor(a, filepath.Join("dir3", "file1")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link5a",
			Origin: repo.PathFor(a, filepath.Join("dir4", "file5")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/link5b",
			Origin: repo.PathFor(a, filepath.Join("dir4", "file5")),
			Type:   "link",
		},
		{
			Layer:  b1.Path(),
			Path:   "dir2/link6a",
			Origin: repo.PathFor(b1, filepath.Join("dir2", "file6")),
			Type:   "link",
		},
		{
			Layer:  b1.Path(),
			Path:   "dir2/link6b",
			Origin: repo.PathFor(b1, filepath.Join("dir2", "file6")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo1a",
			Origin: "github.com/hattya/repo4a",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo1b",
			Origin: "github.com/hattya/repo4b",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo3a_",
			Origin: "github.com/hattya/repo3a",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo3b_",
			Origin: "github.com/hattya/repo3b",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo4a",
			Origin: "github.com/hattya/repo1a",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo4b",
			Origin: "github.com/hattya/repo1b",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo5a",
			Origin: "github.com/hattya/repo5a",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "dir2/repo5b",
			Origin: "github.com/hattya/repo5b",
			Type:   "subrepo",
		},
		{
			Layer:  b1.Path(),
			Path:   "dir2/repo6a",
			Origin: "github.com/hattya/repo6a",
			Type:   "subrepo",
		},
		{
			Layer:  b1.Path(),
			Path:   "dir2/repo6b",
			Origin: "github.com/hattya/repo6b",
			Type:   "subrepo",
		},
		{
			Layer: a.Path(),
			Path:  "dir6/file1",
		},
		{
			Layer: a.Path(),
			Path:  "file1",
		},
		{
			Layer:  a.Path(),
			Path:   "file2_",
			Origin: "file2",
		},
		{
			Layer: b1.Path(),
			Path:  "file3",
		},
		{
			Layer:  a.Path(),
			Path:   "file6",
			Origin: "dir5/file1",
		},
		{
			Layer:  a.Path(),
			Path:   "link1a",
			Origin: repo.PathFor(a, "file1"),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "link1b",
			Origin: repo.PathFor(a, "file1"),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "link2a_",
			Origin: repo.PathFor(a, "file2"),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "link2b_",
			Origin: repo.PathFor(a, "file2"),
			Type:   "link",
		},
		{
			Layer:  b1.Path(),
			Path:   "link3a",
			Origin: repo.PathFor(b1, "file3"),
			Type:   "link",
		},
		{
			Layer:  b1.Path(),
			Path:   "link3b",
			Origin: repo.PathFor(b1, "file3"),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "link6a",
			Origin: repo.PathFor(a, filepath.Join("dir5", "file1")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "link6b",
			Origin: repo.PathFor(a, filepath.Join("dir5", "file1")),
			Type:   "link",
		},
		{
			Layer:  a.Path(),
			Path:   "repo1a",
			Origin: "github.com/hattya/repo1a",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "repo1b",
			Origin: "github.com/hattya/repo1b",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "repo2a_",
			Origin: "github.com/hattya/repo2a",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "repo2b_",
			Origin: "github.com/hattya/repo2b",
			Type:   "subrepo",
		},
		{
			Layer:  b1.Path(),
			Path:   "repo3a",
			Origin: "github.com/hattya/repo3a",
			Type:   "subrepo",
		},
		{
			Layer:  b1.Path(),
			Path:   "repo3b",
			Origin: "github.com/hattya/repo3b",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "repo6a",
			Origin: "github.com/hattya/repo1a",
			Type:   "subrepo",
		},
		{
			Layer:  a.Path(),
			Path:   "repo6b",
			Origin: "github.com/hattya/repo1b",
			Type:   "subrepo",
		},
	}
	if err := wc.SelectLayer(b1.Path()); err != nil {
		t.Fatal(err)
	}
	switch _, err := wc.MergeLayers(); {
	case err != nil:
		t.Error(err)
	case !reflect.DeepEqual(wc.State.WC, e):
		t.Error("unexpected result")
	}

	if err := mkdir(wc.PathFor("dir1")); err != nil {
		t.Fatal(err)
	}
	e[0].Path = "dir1/file1"
	e[0].IsDir = false
	for i := range e {
		if e[i].Layer == b1.Path() {
			e[i].Layer = b2.Path()
			if e[i].Type == "link" {
				e[i].Origin = repo.PathFor(b2, e[i].Origin[len(repo.PathFor(b1, "."))+1:])
			}
		}
	}
	if err := wc.SelectLayer(b2.Path()); err != nil {
		t.Fatal(err)
	}
	switch _, err := wc.MergeLayers(); {
	case err != nil:
		t.Error(err)
	case !reflect.DeepEqual(wc.State.WC, e):
		t.Error("unexpected result")
	}
}

func TestMergeLayersError(t *testing.T) {
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
	a, err := repo.NewLayer("a")
	if err != nil {
		t.Fatal(err)
	}
	b1, err := repo.NewLayer("b/1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.NewLayer("b/2"); err != nil {
		t.Fatal(err)
	}
	// cannot resolve
	if _, err := wc.MergeLayers(); err == nil {
		t.Error("expected error")
	}

	ui := new(testUI)
	git, err := nazuna.VCSFor(ui, filepath.Join(".nzn", "r"))
	if err != nil {
		t.Fatal(err)
	}
	reset := func() {
		a.Links = nil
		a.Subrepos = nil
		b1.Aliases = nil
	}
	alias := func(l *nazuna.Layer, src, dst string) {
		t.Helper()
		if err := l.NewAlias(src, dst); err != nil {
			t.Fatal(err)
		}
	}
	link := func(l *nazuna.Layer, path []string, src, dst string) {
		t.Helper()
		if _, err := l.NewLink(path, src, dst); err != nil {
			t.Fatal(err)
		}
	}
	subrepo := func(l *nazuna.Layer, dst string) {
		t.Helper()
		if _, err := l.NewSubrepo("github.com/hattya/"+filepath.Base(dst), dst); err != nil {
			t.Fatal(err)
		}
	}
	if err := touch(repo.PathFor(a, "file1")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(a, "file2")); err != nil {
		t.Fatal(err)
	}
	if err := git.Add("."); err != nil {
		t.Fatal(err)
	}
	if err := wc.SelectLayer(b1.Path()); err != nil {
		t.Fatal(err)
	}
	// file: file not found
	if err := os.Remove(repo.PathFor(a, "file2")); err != nil {
		t.Fatal(err)
	}
	if _, err := wc.MergeLayers(); err == nil {
		t.Error("expected error")
	}
	if err := touch(repo.PathFor(a, "file2")); err != nil {
		t.Fatal(err)
	}
	// file: alias error
	alias(b1, "file1", filepath.Join("..", "file01"))
	if _, err := wc.MergeLayers(); err == nil {
		t.Error("expected error")
	}
	// link: file not found
	reset()
	link(a, nil, repo.PathFor(a, "file3"), "file03")
	if _, err := wc.MergeLayers(); err != nil {
		t.Error(err)
	}
	// link: alias error
	reset()
	link(a, nil, repo.PathFor(a, "file1"), "file01")
	alias(b1, "file01", filepath.Join("..", "file001"))
	if _, err := wc.MergeLayers(); err == nil {
		t.Error("expected error")
	}
	// link: alias error
	reset()
	link(a, []string{repo.PathFor(a, ".")}, "file1", "file01")
	alias(b1, "file01", filepath.Join("..", "file001"))
	if _, err := wc.MergeLayers(); err == nil {
		t.Error("expected error")
	}
	// link: warning
	reset()
	a.Links = map[string][]*nazuna.Link{
		"": {
			{
				Src: repo.PathFor(a, "file1"),
				Dst: "file1",
			},
		},
	}
	if _, err := wc.MergeLayers(); err != nil {
		t.Error(err)
	}
	// subrepo: alias error
	reset()
	subrepo(a, "repo1")
	alias(b1, "repo1", filepath.Join("..", "repo01"))
	if _, err := wc.MergeLayers(); err == nil {
		t.Error("expected error")
	}
	// subrepo: warning
	reset()
	a.Subrepos = map[string][]*nazuna.Subrepo{
		"": {
			{
				Src:  "github.com/hattya/repo1",
				Name: "file1",
			},
		},
	}
	if _, err := wc.MergeLayers(); err != nil {
		t.Error(err)
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
