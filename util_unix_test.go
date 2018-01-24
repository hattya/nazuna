//
// nazuna :: util_unix_test.go
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

// +build !plan9,!windows

package nazuna_test

import (
	"os"
	"testing"

	"github.com/hattya/nazuna"
)

func TestCreateLink(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(dir)
	}
	defer os.RemoveAll(dir)
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	if err := touch("src"); err != nil {
		t.Error(err)
	}

	if p := "src"; nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = true, expected false", p)
	}
	if err := nazuna.CreateLink("src", "dst"); err != nil {
		t.Error(err)
	}
	if p := "dst"; !nazuna.IsLink("dst") {
		t.Errorf("IsLink(%q) = false, expected true", p)
	}

	if p, o := "dst", "src"; !nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = false, expected true", p, o)
	}
	if p, o := "dst", "_"; nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = true, expected false", p, o)
	}
	if p, o := "src", "dst"; nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = true, expected false", p, o)
	}

	if err := nazuna.Unlink("dst"); err != nil {
		t.Error(err)
	}
	if err := nazuna.Unlink("src"); err == nil {
		t.Error("expected error")
	}
}
