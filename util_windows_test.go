//
// nazuna :: util_windows_test.go
//
//   Copyright (c) 2014-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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
