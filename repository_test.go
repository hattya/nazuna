//
// nazuna :: repository_test.go
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
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestRepository(t *testing.T) {
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

	switch _, err := nazuna.Open(nil, "."); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "no repository found "):
		t.Error("unexpected error:", err)
	}

	if err := mkdir(".nzn/r"); err != nil {
		t.Fatal(err)
	}
	switch _, err = nazuna.Open(nil, "."); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "unknown vcs for directory "):
		t.Error("unexpected error:", err)
	}

	if err := mkdir(".nzn/r/.git"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(".nzn", "r", "nazuna.json")

	if err := mkdir(path); err != nil {
		t.Fatal(err)
	}
	if _, err = nazuna.Open(nil, "."); err == nil {
		t.Error("expected error")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	repo, err := nazuna.Open(nil, ".")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Flush(); err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if g, e := string(data), "[]\n"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

func TestNewLayer(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	layers := []*nazuna.Layer{
		{Name: "a"},
		{Name: "z"},
	}
	if _, err := repo.NewLayer("z"); err != nil {
		t.Error(err)
	}
	if _, err := repo.NewLayer("a"); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(repo.Layers, layers) {
		t.Error("unexpected order")
	}

	repo.Layers = nil
	layers = []*nazuna.Layer{{Name: "abst"}}
	abst := layers[0]
	abst.Layers = []*nazuna.Layer{
		{Name: "a"},
		{Name: "z"},
	}
	abst.SetRepo(repo)
	for _, l := range abst.Layers {
		l.SetAbst(abst)
	}
	if _, err := repo.NewLayer("abst/z"); err != nil {
		t.Error(err)
	}
	if _, err := repo.NewLayer("abst/a"); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(repo.Layers, layers) {
		t.Error("expected to sort by Layer.Name")
	}
}

func TestNewLayerError(t *testing.T) {
	repo, err := init_()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(repo.Root())

	if _, err := repo.NewLayer("layer"); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{
		"",
		"/",
		"//",
		"layer/layer",
		"layer",
	} {
		if _, err := repo.NewLayer(n); err == nil {
			t.Error("expected error")
		}
	}
}

func init_() (repo *nazuna.Repository, err error) {
	dir, err := tempDir()
	if err != nil {
		return
	}
	rdir := filepath.Join(dir, ".nzn", "r")
	if err = mkdir(rdir); err != nil {
		return
	}
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = rdir
	if err = cmd.Run(); err != nil {
		return
	}
	repo, err = nazuna.Open(new(testUI), dir)
	if err != nil {
		return
	}
	return
}
