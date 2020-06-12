//
// nazuna/cmd/nzn :: util.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"fmt"
	"os/exec"

	"github.com/hattya/go.cli"
)

type UI struct {
	*cli.CLI
}

func newUI() *UI {
	return &UI{app}
}

func (ui *UI) Exec(cmd *exec.Cmd) (err error) {
	cmd.Stdin = ui.Stdin
	cmd.Stdout = ui.Stdout
	cmd.Stderr = ui.Stderr
	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("%v: %v", cmd.Args[0], err)
	}
	return err
}
