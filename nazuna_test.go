//
// nazuna :: nazuna_test.go
//
//   Copyright (c) 2013-2025 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hattya/nazuna"
)

func init() {
	nazuna.Discover(false)
}

type testUI struct {
	bytes.Buffer
}

func (*testUI) Print(...any) (int, error)          { return 0, nil }
func (*testUI) Printf(string, ...any) (int, error) { return 0, nil }
func (*testUI) Println(...any) (int, error)        { return 0, nil }
func (*testUI) Error(...any) (int, error)          { return 0, nil }
func (*testUI) Errorf(string, ...any) (int, error) { return 0, nil }
func (*testUI) Errorln(...any) (int, error)        { return 0, nil }

func (ui *testUI) Exec(cmd *exec.Cmd) error {
	cmd.Stdout = ui
	cmd.Stderr = ui
	return cmd.Run()
}

func mkdir(s ...string) error {
	return os.MkdirAll(filepath.Join(s...), 0o777)
}

func sandbox(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("sandbox:", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal("sandbox:", err)
	}
	t.Setenv("PWD", dir)
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Error("sandbox:", err)
		}
	})
	return dir
}

func touch(s ...string) error {
	return os.WriteFile(filepath.Join(s...), []byte{}, 0o666)
}
