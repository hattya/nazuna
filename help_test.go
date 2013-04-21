//
// nazuna :: help_test.go
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
	"testing"

	"github.com/hattya/nazuna"
)

func TestHelp(t *testing.T) {
	rc, bout, berr := runCLI("nazuna.test", "help")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if err := equal(nazuna.HelpOut, bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "--help")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if err := equal(nazuna.HelpOut, bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}
}

func TestHelpUnknownCommand(t *testing.T) {
	rc, bout, berr := runCLI("nazuna.test", "help", "nazuna")
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if err := equal(nazuna.HelpOut, bout); err != nil {
		t.Error(err)
	}
	if berr != "nazuna.test: unknown command 'nazuna'\n" {
		t.Errorf(`expected "nazuna.test: unknown command 'nazuna'\n", got %q`, berr)
	}
}

func TestHelpHelp(t *testing.T) {
	rc, bout, berr := runCLI("nazuna.test", "help", "help")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if err := equal(nazuna.HelpUsage, bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}
}
