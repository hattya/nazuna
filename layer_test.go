//
// nazuna :: layer_test.go
//
//   Copyright (c) 2014-2018 Akinori Hattori <hattya@gmail.com>
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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hattya/nazuna"
)

func TestLayer(t *testing.T) {
	l := &nazuna.Layer{Name: "layer"}
	if g, e := l.Path(), "layer"; g != e {
		t.Errorf("Layer.Path() = %v, expected %v", g, e)
	}

	l.SetAbst(&nazuna.Layer{Name: "abst"})
	if g, e := l.Path(), "abst/layer"; g != e {
		t.Errorf("Layer.Path() = %v, expected %v", g, e)
	}
}

func TestNewAlias(t *testing.T) {
	repo, err := initLayer()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	l, err := repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "dst"); err != nil {
		t.Error(err)
	}

	if err := l.NewAlias("src", "src"); err == nil {
		t.Error("expected error")
	}
	if err := l.NewAlias("src", "dst"); err == nil {
		t.Error("expected error")
	}
}

func TestNewAliasError(t *testing.T) {
	repo, err := initLayer()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	// abstruct layer
	l, err := repo.LayerOf("abst")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "dst"); err == nil {
		t.Error("expected error")
	}
	// already exists: file
	l, err = repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "dst"); err == nil {
		t.Error("expected error")
	}
	// already exists: directory
	l.Aliases = nil
	if err := repo.Command("rm", "-rf", "."); err != nil {
		t.Fatal(err)
	}
	if err := mkdir(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, filepath.Join("dst", "file"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "dst"); err != nil {
		t.Error(err)
	}
}

func TestNewLink(t *testing.T) {
	repo, err := initLayer()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	links := map[string][]*nazuna.Link{
		"": {
			{[]string{"path"}, "src", "dst"},
			{nil, "a", "z"},
		},
	}

	l, err := repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink([]string{"path"}, "src", "dst"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink(nil, "a", "z"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l.Links, links) {
		t.Error("expected to sort by Link.Dst")
	}

	if _, err := l.NewLink([]string{"path"}, "src", "dst"); err == nil {
		t.Error("expected error")
	}
}

func TestNewLinkError(t *testing.T) {
	repo, err := initLayer()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	// abstruct layer
	l, err := repo.LayerOf("abst")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink([]string{}, "src", "dst"); err == nil {
		t.Error("expected error")
	}
	// already exists: file
	l, err = repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink([]string{"path"}, "src", "dst"); err == nil {
		t.Error("expected error")
	}
	// already exists: directory
	l.Links = nil
	if err := repo.Command("rm", "-rf", "."); err != nil {
		t.Fatal(err)
	}
	if err := mkdir(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, filepath.Join("dst", "file"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink([]string{"path"}, "src", "dst"); err == nil {
		t.Error("expected error")
	}
}

func TestNewSubrepo(t *testing.T) {
	repo, err := initLayer()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	src := "github.com/hattya/nazuna"
	subrepos := make(map[string][]*nazuna.Subrepo)
	l, err := repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}

	subrepos[""] = []*nazuna.Subrepo{
		{"a", "z"},
		{src, "dst"},
	}
	if _, err := l.NewSubrepo(src, "dst"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewSubrepo("a", "z"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l.Subrepos, subrepos) {
		t.Error("expected to sort by Subrepo.Src")
	}

	if _, err := l.NewSubrepo(src, "dst"); err == nil {
		t.Error("expected error")
	}

	l.Subrepos = nil
	subrepos[""] = []*nazuna.Subrepo{
		{"a", "z"},
		{src, ""},
	}
	if _, err := l.NewSubrepo(src, filepath.Base(src)); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewSubrepo("a", "z"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l.Subrepos, subrepos) {
		t.Error("expected to sort by Subrepo.Src")
	}

	if _, err := l.NewSubrepo(src, filepath.Base(src)); err == nil {
		t.Error("expected error")
	}
}

func TestNewSubrepoError(t *testing.T) {
	repo, err := initLayer()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	src := "github.com/hattya/nazuna"

	// abstruct layer
	l, err := repo.LayerOf("abst")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewSubrepo(src, "dst"); err == nil {
		t.Error("expected error")
	}
	// already exists: file
	l, err = repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err = l.NewSubrepo(src, "dst"); err == nil {
		t.Error("expected error")
	}
	// already exists: directory
	l.Subrepos = nil
	if err := repo.Command("rm", "-rf", "."); err != nil {
		t.Fatal(err)
	}
	if err := mkdir(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, filepath.Join("dst", "file"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err = l.NewSubrepo(src, "dst"); err == nil {
		t.Error("expected error")
	}
}

func initLayer() (repo *nazuna.Repository, err error) {
	repo, err = init_()
	if err != nil {
		return
	}
	_, err = repo.NewLayer("abst/layer")
	return
}
