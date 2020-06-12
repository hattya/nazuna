//
// nazuna/cmd/nzn :: nzn.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

var app = cli.NewCLI()

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := app.Run(os.Args[1:]); err != nil {
		switch err := err.(type) {
		case cli.FlagError:
			os.Exit(2)
		case cli.Interrupt:
			os.Exit(128 + 2)
		case SystemExit:
			os.Exit(int(err))
		}
		os.Exit(1)
	}
}

func init() {
	app.Version = nazuna.Version
	app.Prepare = prepare
	app.ErrorHandler = errorHandler
}

func prepare(ctx *cli.Context, cmd *cli.Command) error {
	if v, ok := cmd.Data.(bool); ok && v {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		repo, err := nazuna.Open(newUI(), wd)
		if err != nil {
			return err
		}
		ctx.Data = repo
	}
	return nil
}

func errorHandler(ctx *cli.Context, err error) error {
	switch err.(type) {
	case cli.FlagError:
	case cli.Interrupt:
	case SystemExit:
		return err
	default:
		ctx.Stack = nil
	}
	return cli.ErrorHandler(ctx, err)
}

type SystemExit int

func (e SystemExit) Error() string {
	return fmt.Sprintf("exit status %d", e)
}
