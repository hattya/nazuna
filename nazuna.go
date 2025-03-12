//
// nazuna :: nazuna.go
//
//   Copyright (c) 2013-2025 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna

import "os/exec"

const Version = "0.7+"

type UI interface {
	Print(...any) (int, error)
	Printf(string, ...any) (int, error)
	Println(...any) (int, error)
	Error(...any) (int, error)
	Errorf(string, ...any) (int, error)
	Errorln(...any) (int, error)
	Exec(*exec.Cmd) error
}
