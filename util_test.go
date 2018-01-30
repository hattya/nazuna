//
// nazuna :: util_test.go
//
//   Copyright (c) 2018 Akinori Hattori <hattya@gmail.com>
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

func TestIsDir(t *testing.T) {
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

	if !nazuna.IsDir(".") {
		t.Errorf("IsDir(%q) = false, expected true", ".")
	}

	p := "file"
	if err := touch(p); err != nil {
		t.Fatal(err)
	}
	if nazuna.IsDir(p) {
		t.Errorf("IsDir(%q) = true, expected false", p)
	}
}

func TestIsEmptyDir(t *testing.T) {
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

	if !nazuna.IsEmptyDir(".") {
		t.Errorf("IsEmptyDir(%q) = false, expected true", ".")
	}

	p := "file"
	if err := touch(p); err != nil {
		t.Fatal(err)
	}
	if nazuna.IsEmptyDir(p) {
		t.Errorf("IsEmptyDir(%q) = true, expected false", p)
	}
	if nazuna.IsEmptyDir(".") {
		t.Errorf("IsEmptyDir(%q) = true, expected false", ".")
	}
}

func TestSplitPath(t *testing.T) {
	sep := string(os.PathSeparator)
	for _, p := range []string{
		"dir" + sep + "file",
		"dir" + sep + sep + "file",
		"dir/file",
		"dir//file",
	} {
		dir, name := nazuna.SplitPath(p)
		if g, e := []string{dir, name}, []string{"dir", "file"}; !reflect.DeepEqual(g, e) {
			t.Errorf("expected %v, got %v", e, g)
		}
	}
}

func TestSortKeys(t *testing.T) {
	var m interface{}
	m = map[string]string{
		"a": "a",
		"z": "z",
	}
	if g, e := nazuna.SortKeys(m), []string{"a", "z"}; !reflect.DeepEqual(g, e) {
		t.Errorf("expected %v, got %v", e, g)
	}
	// not map
	e := []string(nil)
	if g := nazuna.SortKeys(nil); !reflect.DeepEqual(g, e) {
		t.Errorf("expected %v, got %v", e, g)
	}
	// not map[string]
	m = make(map[int]int)
	if g := nazuna.SortKeys(m); !reflect.DeepEqual(g, e) {
		t.Errorf("expected %v, got %v", e, g)
	}
	m = map[int]int{
		0: 0,
		9: 9,
	}
	if g := nazuna.SortKeys(m); !reflect.DeepEqual(g, e) {
		t.Errorf("expected %v, got %v", e, g)
	}
}

func TestMarshalError(t *testing.T) {
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
	p := filepath.Join(repo.Root(), ".nzn", "r", "nazuna.json")

	if err := nazuna.Marshal(repo, filepath.Base(p), nil); err == nil {
		t.Error("expected error")
	}
	if err := nazuna.Marshal(repo, p, nazuna.Marshal); err == nil {
		t.Error("expected error")
	}
	if err := mkdir(p); err != nil {
		t.Fatal(err)
	}
	if err := nazuna.Marshal(repo, p, nil); err == nil {
		t.Error("expected error")
	}
}

func TestUnmarshalError(t *testing.T) {
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
	p := filepath.Join(repo.Root(), ".nzn", "r", "nazuna.json")

	if err := nazuna.Unmarshal(repo, filepath.Base(p), nil); err == nil {
		t.Error("expected error")
	}
	if err := touch(p); err != nil {
		t.Fatal(err)
	}
	if err := nazuna.Unmarshal(repo, p, nil); err == nil {
		t.Error("expected error")
	}
}
