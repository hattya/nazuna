//
// nazuna :: layer_test.go
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
	"strings"
	"testing"
)

func TestLayer(t *testing.T) {
	dir, err := ioutil.TempDir("", "nazuna.test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rc, _, berr := runCLI("nazuna.test", "init", "--vcs=git")
	if rc != 0 {
		t.Logf("stderr:\n%s", berr)
		t.Fatalf("expected 0, got %d", rc)
	}

	rc, bout, berr := runCLI("nazuna.test", "layer")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "layer", "-c", "a")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "layer", "-c", "b")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "layer")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if err := equal("b\na\n", bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "layer", "-c", "a")
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if !strings.Contains(berr, ": layer 'a' already exists!") {
		t.Error("error expected")
	}
}

func TestLayerError(t *testing.T) {
	dir, err := ioutil.TempDir("", "nazuna.test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rc, bout, berr := runCLI("nazuna.test", "layer")
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if !strings.Contains(berr, ": no repository found ") {
		t.Error("error expected")
	}

	rc, _, berr = runCLI("nazuna.test", "init", "--vcs=git")
	if rc != 0 {
		t.Logf("stderr:\n%s", berr)
		t.Fatalf("expected 0, got %d", rc)
	}
	rc, bout, berr = runCLI("nazuna.test", "layer", "-c")
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if !strings.Contains(berr, ": invalid arguments") {
		t.Error("error expected")
	}
}
