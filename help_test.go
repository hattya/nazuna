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
	"fmt"
	"testing"
)

const (
	helpUsage = `usage: nzn help [options] [--] [command]

  display help information about nazuna

`
	helpOut = `nazuna - A layered dotfiles management

list of commands:

  clone      make a copy of an existing repository
  help       display help information about nazuna
  init       create a new repository in the specified directory
  layer      manage repository layers
  link       create a link for the specified path
  update     update working copy
  vcs        run the vcs command inside the repository
  version    output version and copyright information

`
)

func TestHelp(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"nzn", "help"},
			out: helpOut,
		},
		{
			cmd: []string{"nzn", "help", "help"},
			out: helpUsage,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestHelpError(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"nzn", "help", "nazuna"},
			out: fmt.Sprintf("nzn: unknown command 'nazuna'\n%s[1]\n", helpOut),
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}
