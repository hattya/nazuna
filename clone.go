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
	Usage: "clone --vcs=<type> <repository> [<path>]",
	Help: `
  make a copy of an existing repository

options:

      --vcs=<type>    vcs type
`,
}

var cloneVCS string

func init() {
	cmdClone.Run = runClone
	cmdClone.Flag.StringVar(&cloneVCS, "vcs", "", "")
}

func runClone(ui UI, args []string) error {
	if len(args) == 0 {
		return errArg
	}
	src := args[0]

	rootdir := "."
	if 1 < len(args) {
		rootdir = args[1]
	}
	nzndir := filepath.Join(rootdir, ".nzn")
	if !isEmptyDir(nzndir) {
		return fmt.Errorf("repository '%s' already exists!", rootdir)
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
	return ui.Exec(vcs.Clone(src, filepath.Join(nzndir, "repo")))
}
