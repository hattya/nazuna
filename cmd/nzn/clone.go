//
// nazuna/cmd/nzn :: clone.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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
		Name:  []string{"clone"},
		Usage: "--vcs <type> <repository> [<path>]",
		Desc: strings.TrimSpace(cli.Dedent(`
			create a copy of an existing repository

			  Create a copy of an existing repository in <path>. If <path> does not exist,
			  it will be created.

			  If <path> is not specified, the current working diretory is used.
		`)),
		Flags:  flags,
		Action: clone,
	})
}

func clone(ctx *cli.Context) error {
	if len(ctx.Args) == 0 {
		return cli.ErrArgs
	}
	src := ctx.Args[0]

	root := "."
	if 1 < len(ctx.Args) {
		root = ctx.Args[1]
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
	return vcs.Clone(src, filepath.Join(nzndir, "r"))
}
