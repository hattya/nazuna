//
// nazuna :: nazuna_test.go
//
//   Copyright (c) 2013-2018 Akinori Hattori <hattya@gmail.com>
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

package nazuna_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hattya/nazuna"
)

var entryTests = []struct {
	e *nazuna.Entry
	s string
}{
	{
		&nazuna.Entry{},
		"!",
	},
	{
		&nazuna.Entry{
			IsDir: true,
		},
		"!",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
		},
		"!layer",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
			IsDir: true,
		},
		"!layer",
	},
	{
		&nazuna.Entry{
			Path: "path",
		},
		"path!",
	},
	{
		&nazuna.Entry{
			Path:  "path",
			IsDir: true,
		},
		"path/!",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
			Path:  "path",
		},
		"path!layer",
	},
	{
		&nazuna.Entry{
			Layer: "layer",
			Path:  "path",
			IsDir: true,
		},
		"path/!layer",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
		},
		"path!layer:origin",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
			IsDir:  true,
		},
		"path/!layer:origin/",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
			Type:   "link",
		},
		"path!origin",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "origin",
			IsDir:  true,
			Type:   "link",
		},
		"path/!origin" + string(os.PathSeparator),
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "github.com/hattya/nazuna",
			Type:   "subrepo",
		},
		"path!github.com/hattya/nazuna",
	},
	{
		&nazuna.Entry{
			Layer:  "layer",
			Path:   "path",
			Origin: "github.com/hattya/nazuna",
			IsDir:  true,
			Type:   "subrepo",
		},
		"path/!github.com/hattya/nazuna",
	},
}

func TestEntry(t *testing.T) {
	for _, tt := range entryTests {
		if g, e := tt.e.Format("%v!%v"), tt.s; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
}

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
	return popd, os.Chdir(path)
}

func tempDir() (string, error) {
	return ioutil.TempDir("", "nazuna.test")
}

func touch(s ...string) error {
	return ioutil.WriteFile(filepath.Join(s...), []byte{}, 0666)
}
