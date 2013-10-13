//
// nazuna :: clone.go
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

package nazuna

import (
	"fmt"
	"os"
	"path/filepath"
)

var cmdClone = &Command{
	Names: []string{"clone"},
	Usage: []string{
		"clone --vcs <type> <repository> [<path>]",
	},
	Help: `
create a copy of an existing repository

  Create a copy of an existing repository in <path>. If <path> does not exist,
  it will be created.

  If <path> is not specified, the current working diretory is used.

options:

      --vcs <type>    vcs type
`,
}

var cloneVCS string

func init() {
	cmdClone.Flag.StringVar(&cloneVCS, "vcs", "", "")

	cmdClone.Run = runClone
}

func runClone(ui UI, args []string) error {
	if len(args) == 0 {
		return ErrArg
	}
	src := args[0]

	root := "."
	if 1 < len(args) {
		root = args[1]
	}
	nzndir := filepath.Join(root, ".nzn")
	if !isEmptyDir(nzndir) {
		return fmt.Errorf("repository '%s' already exists!", root)
	}

	if cloneVCS == "" {
		return FlagError("flag --vcs is required")
	}
	vcs, err := FindVCS(cloneVCS)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(nzndir, 0777); err != nil {
		return err
	}
	return ui.Exec(vcs.Clone(src, filepath.Join(nzndir, "r")))
}
