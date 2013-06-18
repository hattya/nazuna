//
// nazuna :: nazuna_test.go
//
//   Copyright (c) 2013 Akinori Hattori <hattya@gmail.com>
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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hattya/nazuna"
	"github.com/mb0/diff"
)

var testVarRe = regexp.MustCompile(`\$[[:alnum:]]+`)
var testEnv []string

func init() {
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "LANG") || strings.HasPrefix(v, "LC_") {
			continue
		}
		testEnv = append(testEnv, v)
	}
}

type testCmdLine struct {
	cmd []string
	out string
}

type testScript []*testCmdLine

func (t testScript) run() error {
	vars := make(map[string]string)
	for i, c := range t {
		errorf := func(a interface{}) error {
			var s string
			switch v := a.(type) {
			case string:
				s = v
			case error:
				s = v.Error()
			}
			return t.errorf(i+1, c, s)
		}
		args := t.expand(c.cmd[1:], vars)
		switch c.cmd[0] {
		case "cat":
			data, err := ioutil.ReadFile(args[0])
			if err != nil {
				return errorf(err)
			}
			if err := errorf(t.diff(c.out, string(data), 0)); err != nil {
				return err
			}
		case "cd":
			popd, err := pushd(args[0])
			if err != nil {
				return errorf(err)
			}
			defer popd()
		case "git":
			cmd := exec.Command(c.cmd[0], args...)
			cmd.Env = testEnv
			b, err := cmd.CombinedOutput()
			out := string(b)
			rc := 0
			if err != nil {
				out += fmt.Sprintf("%s\n", err)
				rc = 1
			}
			if err := errorf(t.diff(c.out, out, rc)); err != nil {
				return err
			}
		case "ln":
			if len(args) != 3 || args[0] != "-s" {
				return errorf("ln: invalid arguments")
			}
			if err := nazuna.Ln(args[1], args[2]); err != nil {
				return errorf(err)
			}
		case "ls":
			f, err := os.Open(args[0])
			if os.IsNotExist(err) {
				return errorf(err)
			}
			defer f.Close()
			list, err := f.Readdir(-1)
			if err != nil {
				return errorf(err)
			}
			out := new(bytes.Buffer)
			rc := 0
			for _, fi := range list {
				var s string
				switch {
				case fi.Mode().IsRegular():
				case fi.Mode().IsDir():
					s = "/"
				case fi.Mode()&os.ModeSymlink != 0:
				default:
					s = ">"
					rc = 1
				}
				fmt.Fprintf(out, "%s%s\n", fi.Name(), s)
			}
			if err := errorf(t.diff(c.out, out.String(), rc)); err != nil {
				return err
			}
		case "mkdir":
			if err := mkdir(args[0]); err != nil {
				return errorf(err)
			}
		case "mkdtemp":
			dir, err := mkdtemp()
			if err != nil {
				return errorf(err)
			}
			defer nazuna.RemoveAll(dir)
			vars["$tempdir"] = dir
		case "nzn":
			for _, c := range nazuna.Commands {
				c.Flag.Visit(func(f *flag.Flag) {
					c.Flag.Set(f.Name, f.DefValue)
				})
			}
			out := new(bytes.Buffer)
			nzn := nazuna.NewCLI(append([]string{c.cmd[0]}, args...))
			nzn.SetOut(out)
			nzn.SetErr(out)
			rc := nzn.Run()
			if err := errorf(t.diff(c.out, out.String(), rc)); err != nil {
				return err
			}
		case "rm":
			var remove func(string) error
			switch {
			case 1 < len(args) && args[0] == "-r":
				remove = nazuna.RemoveAll
				args = args[1:]
			default:
				remove = os.Remove
			}
			if err := remove(args[0]); err != nil {
				return errorf(err)
			}
		case "touch":
			if err := touch(args[0]); err != nil {
				return errorf(err)
			}
		default:
			return errorf(fmt.Sprintf("command not found: %s", c.cmd[0]))
		}
	}
	return nil
}

func (t testScript) diff(aout, bout string, rc int) string {
	if 0 < len(bout) && bout[len(bout)-1] != '\n' {
		bout += " (no-eol)\n"
	}
	if rc != 0 {
		bout += fmt.Sprintf("[%d]\n", rc)
	}
	a := strings.Split(strings.TrimSuffix(aout, "\n"), "\n")
	b := strings.Split(strings.TrimSuffix(bout, "\n"), "\n")
	lno := 0
	buf := new(bytes.Buffer)
	equal := func(i, j int) {
		for ; i < j; i++ {
			fmt.Fprintf(buf, " %s\n", a[i])
		}
	}
	insert := func(i, j int) {
		for ; i < j; i++ {
			fmt.Fprintf(buf, "+%s\n", b[i])
		}
	}
	if aout == "" {
		if bout != "" {
			insert(0, len(b))
		}
	} else {
		cl := diff.Diff(len(a), len(b), &lines{a, b})
		if 0 < len(cl) {
			for _, c := range cl {
				equal(lno, c.A)
				for lno = c.A; lno < c.A+c.Del; lno++ {
					fmt.Fprintf(buf, "-%s\n", a[lno])
				}
				insert(c.B, c.B+c.Ins)
			}
			equal(lno, len(a))
		}
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func (t testScript) errorf(i int, c *testCmdLine, s string) error {
	if s == "" {
		return nil
	}
	return fmt.Errorf("script:%d:\n$ %s\n%s", i, strings.Join(c.cmd, " "), s)
}

func (t testScript) expand(args []string, vars map[string]string) []string {
	list := make([]string, len(args))
	for i, a := range args {
		list[i] = testVarRe.ReplaceAllStringFunc(a, func(s string) string {
			if s, ok := vars[s]; ok {
				return s
			}
			return s
		})
	}
	return list
}

type lines struct {
	a []string
	b []string
}

func (d *lines) Equal(i, j int) bool {
	if strings.HasSuffix(d.a[i], " (re)") {
		m, err := regexp.MatchString(d.a[i][:len(d.a[i])-5], d.b[j])
		if err != nil || !m {
			return false
		}
		return true
	}
	return d.a[i] == d.b[j]
}

func pushd(path string) (func(), error) {
	wd, err := os.Getwd()
	popd := func() {
		if !os.IsNotExist(err) {
			os.Chdir(wd)
		}
	}
	return popd, os.Chdir(path)
}

func mkdtemp() (string, error) {
	return ioutil.TempDir("", "nazuna.test")
}

func mkdir(a ...string) error {
	return os.MkdirAll(filepath.Join(a...), 0777)
}

func touch(a ...string) error {
	return ioutil.WriteFile(filepath.Join(a...), []byte{}, 0666)
}
