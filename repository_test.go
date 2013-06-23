//
// nazuna :: repository_test.go
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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestRepository(t *testing.T) {
	dir, err := mkdtemp()
	if err != nil {
		t.Fatal(err)
	}
	defer nazuna.RemoveAll(dir)

	if _, err := nazuna.OpenRepository(nil, dir); !strings.HasPrefix(err.Error(), "no repository found ") {
		t.Error("error expected")
	}

	if err := mkdir(dir, ".nzn", "r"); err != nil {
		t.Fatal(err)
	}
	if _, err = nazuna.OpenRepository(nil, dir); !strings.HasPrefix(err.Error(), "unknown vcs for directory ") {
		t.Error(err)
	}

	if err := mkdir(dir, ".nzn", "r", ".git"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".nzn", "r", "nazuna.json")

	if err := mkdir(path); err != nil {
		t.Fatal(err)
	}
	if _, err = nazuna.OpenRepository(nil, dir); err == nil {
		t.Error("error expected")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	repo, err := nazuna.OpenRepository(nil, dir)
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
	if string(data) != "[]\n" {
		t.Errorf(`expected "[]\n", got %q`, data)
	}
}
