//
// nazuna :: util_windows_test.go
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

package nazuna_test

import (
	"os"
	"testing"

	"github.com/hattya/nazuna"
)

func TestCreateLink(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer nazuna.RemoveAll(dir)
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	// hardlink
	if err := nazuna.CreateLink("src\x00", "dst"); err == nil {
		t.Error("expected error")
	}
	if err := nazuna.CreateLink("src", "dst\x00"); err == nil {
		t.Error("expected error")
	}
	if err := nazuna.CreateLink("src", "dst"); err == nil {
		t.Error("expected error")
	}
	if g, e := nazuna.IsLink("src\x00"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := nazuna.IsLink("src"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := nazuna.LinksTo("src", "dst"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}

	if err := touch("src"); err != nil {
		t.Fatal(err)
	}
	if g, e := nazuna.IsLink("src"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if err := nazuna.CreateLink("src", "dst"); err != nil {
		t.Error(err)
	}
	if g, e := nazuna.IsLink("src"), true; g != e {
		t.Errorf("expected %v, got %v", e, g)
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
	if err := nazuna.Unlink("dst"); err != nil {
		t.Error(err)
	}
	if err := nazuna.Unlink("src"); err == nil {
		t.Error("expected error")
	}

	// junction
	if err := mkdir("srcdir"); err != nil {
		t.Fatal(err)
	}
	if err := touch("dstdir"); err != nil {
		t.Fatal(err)
	}
	if err := nazuna.CreateLink("srcdir", "dstdir"); err == nil {
		t.Error("expected error")
	}
	if g, e := nazuna.LinksTo("dstdir", "srcdir"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if err := os.Remove("dstdir"); err != nil {
		t.Fatal(err)
	}

	if err := nazuna.CreateLink("srcdir", "dstdir"); err != nil {
		t.Error(err)
	}
	if g, e := nazuna.IsLink("srcdir"), false; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := nazuna.IsLink("dstdir"), true; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := nazuna.LinksTo("dstdir", "srcdir"), true; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if err := nazuna.Unlink("dstdir"); err != nil {
		t.Error(err)
	}
	if err := nazuna.Unlink("srcdir"); err == nil {
		t.Error("expected error")
	}
}
