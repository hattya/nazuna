//
// nazuna :: util_windows_test.go
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
	"testing"

	"github.com/hattya/nazuna"
)

func TestCreateLink(t *testing.T) {
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
	if p := "src\x00"; nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = true, expected false", p)
	}
	if p := "src"; nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = true, expected false", p)
	}
	if p, o := "src", "dst"; nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = true, expected false", p, o)
	}

	if err := touch("src"); err != nil {
		t.Fatal(err)
	}
	if p := "src\x00"; nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = true, expected false", p)
	}
	if err := nazuna.CreateLink("src", "dst"); err != nil {
		t.Error(err)
	}
	if p := "src"; !nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = false, expected true", p)
	}
	if p := "dst"; !nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = false, expected true", p)
	}
	if p, o := "dst", "src"; !nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = false, expected true", p, o)
	}
	if p, o := "dst", "_"; nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = true, expected false", p, o)
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
	if p, o := "dstdir", "srcdir"; nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = true, expected false", p, o)
	}
	if err := os.Remove("dstdir"); err != nil {
		t.Fatal(err)
	}

	if err := nazuna.CreateLink("srcdir", "dstdir"); err != nil {
		t.Error(err)
	}
	if p := "srcdir"; nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = true, expected false", p)
	}
	if p := "dstdir"; !nazuna.IsLink(p) {
		t.Errorf("IsLink(%q) = false, expected true", p)
	}
	if p, o := "dstdir", "srcdir"; !nazuna.LinksTo(p, o) {
		t.Errorf("LinksTo(%q, %q) = false, expected true", p, o)
	}
	if err := nazuna.Unlink("dstdir"); err != nil {
		t.Error(err)
	}
	if err := nazuna.Unlink("srcdir"); err == nil {
		t.Error("expected error")
	}
}
