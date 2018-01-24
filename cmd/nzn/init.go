//
// nazuna/cmd/nzn :: init.go
//
//   Copyright (c) 2013-2018 Akinori Hattori <hattya@gmail.com>
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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

func init() {
	flags := cli.NewFlagSet()
	flags.String("vcs", "", "vcs type")
	flags.MetaVar("vcs", " <type>")

	app.Add(&cli.Command{
		Name:  []string{"init"},
		Usage: "--vcs <type> [<path>]",
		Desc: strings.TrimSpace(cli.Dedent(`
			create a new repository in the specified directory

			  Create a new repository in <path>. If <path> does not exist, it will be
			  created.

			  If <path> is not specified, the current working diretory is used.
		`)),
		Flags:  flags,
		Action: init_,
	})
}

func init_(ctx *cli.Context) error {
	root := "."
	if 0 < len(ctx.Args) {
		root = ctx.Args[0]
	}
	nzndir := filepath.Join(root, ".nzn")
	if !nazuna.IsEmptyDir(nzndir) {
		return fmt.Errorf("repository '%v' already exists!", root)
	}

	if ctx.String("vcs") == "" {
		return cli.FlagError("--vcs flag is required")
	}
	ui := newUI()
	vcs, err := nazuna.FindVCS(ui, ctx.String("vcs"), "")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(nzndir, 0777); err != nil {
		return err
	}
	if err := vcs.Init(filepath.Join(nzndir, "r")); err != nil {
		return err
	}

	repo, err := nazuna.Open(ui, root)
	if err != nil {
		return err
	}
	if err := repo.Flush(); err != nil {
		return err
	}
	return repo.Add(".")
}
