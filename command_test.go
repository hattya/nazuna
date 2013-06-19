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

	if _, err := nazuna.FindCommand(list, "clone"); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Error("error expected")
	}
	if _, err := nazuna.FindCommand(list, "s"); err == nil || !strings.Contains(err.Error(), " ambiguous:") {
		t.Error("error expected")
	}

	switch cmd, err := nazuna.FindCommand(list, "st"); {
	case err != nil:
		t.Error(err)
	case cmd.Name() != "status":
		t.Errorf(`expected "status", got %q`, cmd.Name())
	}
	switch cmd, err := nazuna.FindCommand(list, "stash"); {
	case err != nil:
		t.Error(err)
	case cmd.Name() != "stash":
		t.Errorf(`expected "stash", got %q`, cmd.Name())
	}
}

func TestSortCommandsByName(t *testing.T) {
	list := []*nazuna.Command{
		{Names: []string{"z"}},
		{Names: []string{"a"}},
		{},
	}
	sorted := nazuna.SortCommands(list)

	if list[0].Name() != "z" {
		t.Errorf(`expected "z", got "%s"`, list[0].Name())
	}
	if list[1].Name() != "a" {
		t.Errorf(`expected "a", got "%s"`, list[1].Name())
	}
	if list[2].Name() != "" {
		t.Errorf(`expected "", got "%s"`, list[2].Name())
	}

	if sorted[0].Name() != "" {
		t.Errorf(`expected "", got "%s"`, sorted[0].Name())
	}
	if sorted[1].Name() != "a" {
		t.Errorf(`expected "a", got "%s"`, sorted[1].Name())
	}
	if sorted[2].Name() != "z" {
		t.Errorf(`expected "z", got "%s"`, sorted[2].Name())
	}
}
