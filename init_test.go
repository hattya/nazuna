//
// nazuna :: init_test.go
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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestInit(t *testing.T) {
	dir, err := mkdtemp()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	rc, bout, berr := runCLI("nazuna.test", "init", "--vcs=git", dir)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	path := filepath.Join(".nzn", "repo", ".git")
	fi, err := os.Stat(filepath.Join(dir, path))
	switch {
	case err != nil:
		t.Error(err)
	case !fi.Mode().IsDir():
		t.Errorf("%q is not a directory", path)
	}

	path = filepath.Join(".nzn", "repo", "nazuna.json")
	fi, err = os.Stat(filepath.Join(dir, path))
	switch {
	case err != nil:
		t.Fatal(err)
	case !fi.Mode().IsRegular():
		t.Fatalf("%q is not a file", path)
	}

	var layers []*nazuna.Layer
	data, err := ioutil.ReadFile(filepath.Join(dir, path))
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(data, &layers)
	switch {
	case err != nil:
		t.Error(err)
	case len(layers) != 0:
		t.Errorf("expected 0, got %d", len(layers))
	}
}

func TestInitError(t *testing.T) {
	dir, err := mkdtemp()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	rc, bout, berr := runCLI("nazuna.test", "init", "--vcs=cvs", dir)
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if !strings.Contains(berr, ": unknown vcs 'cvs'") {
		t.Errorf("error expected")
	}

	rc, bout, berr = runCLI("nazuna.test", "init", dir)
	if rc != 2 {
		t.Errorf("expected 2, got %d", rc)
	}
	if err := equal(nazuna.InitUsage, bout); err != nil {
		t.Error(err)
	}
	if !strings.Contains(berr, ": flag --vcs is required") {
		t.Errorf("error expected")
	}

	if err := mkdir(dir, ".nzn", "repo"); err != nil {
		t.Fatal(err)
	}
	rc, bout, berr = runCLI("nazuna.test", "init", "--vcs=git", dir)
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if !strings.Contains(berr, " already exists!") {
		t.Errorf("error expected")
	}
}
