//
// nazuna/cmd/nzn :: vcs.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"strings"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

func init() {
	app.Add(&cli.Command{
		Name:  []string{"vcs"},
		Usage: "[<args>]",
		Desc: strings.TrimSpace(cli.Dedent(`
			run the vcs command inside the repository
		`)),
		Action: vcs,
		Data:   true,
	})
}

func vcs(ctx *cli.Context) error {
	repo := ctx.Data.(*nazuna.Repository)
	return repo.Command(ctx.Args...)
}
