//
// nzn :: nzn.go
//
//   Copyright (c) 2013-2014 Akinori Hattori <hattya@gmail.com>
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
	"runtime"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

var app = cli.NewCLI()

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app.Version = nazuna.Version
	app.Prepare = prepare
	app.ErrorHandler = errorHandler

	if err := app.Run(os.Args[1:]); err != nil {
		switch err := err.(type) {
		case cli.FlagError:
			os.Exit(2)
		case SystemExit:
			os.Exit(int(err))
		}
		os.Exit(1)
	}
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
