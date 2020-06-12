//
// nazuna/cmd/nzn :: version.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import "github.com/hattya/go.cli"

func init() {
	cli.Version = showVersion

	app.Add(cli.NewVersionCommand())
}

func showVersion(ctx *cli.Context) error {
	ctx.UI.Printf("nzn version %v\n", ctx.UI.Version)
	return nil
}
