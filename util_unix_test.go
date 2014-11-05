//
// nazuna :: util_unix_test.go
//
//   Copyright (c) 2014 Akinori Hattori <hattya@gmail.com>
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
	"testing"

	"github.com/hattya/nazuna"
)

func TestCreateLink(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(dir)
	}
	defer nazuna.RemoveAll(dir)
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	if err := touch("src"); err != nil {
		t.Error(err)
	}

	if g, e := nazuna.IsLink("src"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if err := nazuna.CreateLink("src", "dst"); err != nil {
		t.Error(err)
	}
	if g, e := nazuna.IsLink("dst"), true; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}

	if g, e := nazuna.LinksTo("dst", "src"), true; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := nazuna.LinksTo("dst", "_"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := nazuna.LinksTo("src", "dst"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}

	if err := nazuna.Unlink("dst"); err != nil {
		t.Error(err)
	}
	if err := nazuna.Unlink("src"); err == nil {
		t.Error("expected error")
	}
}
