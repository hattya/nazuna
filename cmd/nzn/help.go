//
// nazuna/cmd/nzn :: help.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import "github.com/hattya/go.cli"

func init() {
	cli.Usage = formatUsage

	app.Add(cli.NewHelpCommand())
}

func formatUsage(ctx *cli.Context) []string {
	if len(ctx.Stack) == 0 {
		return []string{"Nazuna - A layered dotfiles management"}
	}
	return cli.FormatUsage(ctx)
}
