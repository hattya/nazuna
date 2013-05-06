//
// nazuna :: init.go
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
	"path/filepath"
)

var cmdInit = &Command{
	Names: []string{"init"},
	Usage: "init --vcs=<type> [<path>]",
	Help: `
  create a new repository in the specified directory

options:

      --vcs=<type>    vcs type
`,
}

var initVCS string

func init() {
	cmdInit.Run = runInit
	cmdInit.Flag.StringVar(&initVCS, "vcs", "", "")
}

func runInit(ui UI, args []string) error {
	rootdir := "."
	if 0 < len(args) {
		rootdir = args[0]
	}
	nzndir := filepath.Join(rootdir, ".nzn")
	if !isEmptyDir(nzndir) {
		return fmt.Errorf("repository '%s' already exists!", rootdir)
	}

	if initVCS == "" {
		return FlagError("flag --vcs is required")
	}
	vcs, err := FindVCS(initVCS)
	if err != nil {
		return err
	}
	if err := ui.Exec(vcs.Init(filepath.Join(nzndir, "repo"))); err != nil {
		return err
	}

	repo, err := OpenRepository(ui, rootdir)
	if err != nil {
		return err
	}
	err = repo.Flush()
	if err != nil {
		return err
	}
	return repo.Add(".")
}
