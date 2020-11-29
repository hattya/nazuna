//
// nazuna :: nazuna_test.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hattya/nazuna"
)

func init() {
	nazuna.Discover(false)
}

type testUI struct {
	bytes.Buffer
}

func (*testUI) Print(...interface{}) (int, error)          { return 0, nil }
func (*testUI) Printf(string, ...interface{}) (int, error) { return 0, nil }
func (*testUI) Println(...interface{}) (int, error)        { return 0, nil }
func (*testUI) Error(...interface{}) (int, error)          { return 0, nil }
func (*testUI) Errorf(string, ...interface{}) (int, error) { return 0, nil }
func (*testUI) Errorln(...interface{}) (int, error)        { return 0, nil }

func (ui *testUI) Exec(cmd *exec.Cmd) error {
	cmd.Stdout = ui
	cmd.Stderr = ui
	return cmd.Run()
}

func mkdir(s ...string) error {
	return os.MkdirAll(filepath.Join(s...), 0777)
}

func pushd(path string) (func() error, error) {
	wd, err := os.Getwd()
	popd := func() error {
		if err == nil {
			return os.Chdir(wd)
		}
		return err
	}
	os.Setenv("PWD", path)
	return popd, os.Chdir(path)
}

func tempDir() (string, error) {
	return ioutil.TempDir("", "nazuna.test")
}

func touch(s ...string) error {
	return ioutil.WriteFile(filepath.Join(s...), []byte{}, 0666)
}
