//
// nazuna :: util_unix_test.go
//
//   Copyright (c) 2014-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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
