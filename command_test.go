//
// nazuna :: command_test.go
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
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestFindCommand(t *testing.T) {
	run := func(_ nazuna.UI, _ []string) error { return nil }
	list := []*nazuna.Command{
		{
			Names: []string{"add"},
			Run:   run,
		},
		{
			Names: []string{"status", "st"},
			Run:   run,
		},
		{
			Names: []string{"stash"},
			Run:   run,
		},
	}

	switch _, err := nazuna.FindCommand(list, "clone"); {
	case err == nil:
		t.Error("expected error")
	case !strings.Contains(err.Error(), "unknown command"):
		t.Error("unexpected error:", err)
	}
	switch _, err := nazuna.FindCommand(list, "s"); {
	case err == nil:
		t.Error("expected error")
	case !strings.Contains(err.Error(), " ambiguous:"):
		t.Error("unexpected error:", err)
	}

	cmd, err := nazuna.FindCommand(list, "st")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := cmd.Name(), "status"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	cmd, err = nazuna.FindCommand(list, "stash")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := cmd.Name(), "stash"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

func TestSortCommandsByName(t *testing.T) {
	list := []*nazuna.Command{
		{Names: []string{"z"}},
		{Names: []string{"a"}},
		{},
	}
	sorted := nazuna.SortCommands(list)

	if g, e := list[0].Name(), "z"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := list[1].Name(), "a"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := list[2].Name(), ""; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	if g, e := sorted[0].Name(), ""; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := sorted[1].Name(), "a"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := sorted[2].Name(), "z"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}
