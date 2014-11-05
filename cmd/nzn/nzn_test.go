//
// nzn :: nzn_test.go
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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hattya/go.cli"
	"github.com/hattya/go.diff"
	"github.com/hattya/nazuna"
)

func init() {
	nazuna.Discover(false)

	app.Name = "nzn"
	app.Version = nazuna.Version
	app.Prepare = prepare
	app.ErrorHandler = errorHandler
}

type shell struct {
	dir         string
	env         map[string]string
	gitconfig   map[string]string
	funcs       map[string]interface{}
	atexitFuncs []func()
}

func newShell() (*shell, error) {
	dir, err := ioutil.TempDir("", "nzn.test")
	if err != nil {
		return nil, err
	}
	sh := &shell{
		dir: dir,
		env: map[string]string{
			"tempdir": dir,
		},
		gitconfig: map[string]string{
			"core.autocrlf": "false",
			"user.email":    "nazuna@example.com",
		},
	}
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "LANG") || strings.HasPrefix(v, "LC_") || strings.Contains(v, "PWD=") {
			continue
		}
		i := strings.Index(v, "=")
		if i == -1 {
			continue
		}
		sh.env[v[:i]] = v[i+1:]
	}
	sh.funcs = map[string]interface{}{
		"cat":    sh.cat,
		"cd":     sh.cd,
		"export": sh.export,
		"git":    sh.git,
		"ln":     sh.ln,
		"ls":     sh.ls,
		"mkdir":  sh.mkdir,
		"nzn":    sh.nzn,
		"rm":     sh.rm,
		"setup":  sh.setup,
		"touch":  sh.touch,
	}
	sh.atexit(func() { nazuna.RemoveAll(sh.dir) })
	return sh, nil
}

func (sh *shell) run(s script) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err = os.Chdir(sh.dir); err != nil {
		return err
	}
	defer os.Chdir(wd)

	i := 1
	for _, c := range s {
		args := sh.expand(c.cmd[1:]...)
		out, rc := "unknown command", 1
		switch f := sh.funcs[c.cmd[0]].(type) {
		case func(*shell, ...string) (string, int):
			out, rc = f(sh, args...)
		case func(...string) (string, int):
			out, rc = f(args...)
		}
		if diff := sh.verify(c.out, out, rc); diff != "" {
			return fmt.Errorf("script:%d:\n$ %v\n%v", i, strings.Join(c.cmd, " "), diff)
		}
		i++
	}
	return nil
}

func (sh *shell) expand(args ...string) []string {
	list := make([]string, len(args))
	for i, a := range args {
		list[i] = os.Expand(a, func(k string) string {
			if 'a' <= k[0] && k[0] <= 'z' {
				if v, ok := sh.env[k]; ok {
					return v
				}
			}
			return "$" + k
		})
	}
	return list
}

func (sh shell) report(err error) (string, int) {
	if err != nil {
		return err.Error(), 1
	}
	return "", 0
}

func (sh shell) verify(aout, bout string, rc int) string {
	if bout != "" && bout[len(bout)-1] != '\n' {
		bout += " (no-eol)\n"
	}
	if rc != 0 {
		bout += fmt.Sprintf("[%d]\n", rc)
	}
	a := strings.Split(strings.TrimSuffix(aout, "\n"), "\n")
	b := strings.Split(strings.TrimSuffix(bout, "\n"), "\n")
	var buf bytes.Buffer
	format := func(sign string, lines []string, i, j int) {
		for ; i < j; i++ {
			fmt.Fprintf(&buf, "%v%v\n", sign, lines[i])
		}
	}
	switch {
	case aout == "":
		if bout != "" {
			format("+", b, 0, len(b))
		}
	case bout == "":
		format("-", a, 0, len(a))
	default:
		cl := diff.Diff(len(a), len(b), &lines{a, b})
		if 0 < len(cl) {
			lno := 0
			for _, c := range cl {
				format(" ", a, lno, c.A)
				format("-", a, c.A, c.A+c.Del)
				format("+", b, c.B, c.B+c.Ins)
				lno = c.A + c.Del
			}
			format(" ", a, lno, len(a))
		}
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func (sh *shell) atexit(f func()) {
	sh.atexitFuncs = append(sh.atexitFuncs, f)
}

func (sh *shell) exit() {
	for i := len(sh.atexitFuncs) - 1; 0 <= i; i-- {
		sh.atexitFuncs[i]()
	}
}

func (sh *shell) cat(args ...string) (string, int) {
	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		return sh.report(err)
	}
	return string(data), 0
}

func (sh *shell) cd(args ...string) (string, int) {
	return sh.report(os.Chdir(args[0]))
}

func (sh *shell) export(args ...string) (string, int) {
	kv := strings.SplitN(args[0], "=", 2)
	v := os.Getenv(kv[0])
	if err := os.Setenv(kv[0], kv[1]); err != nil {
		return sh.report(err)
	}
	sh.atexit(func() { os.Setenv(kv[0], v) })
	sh.env[kv[0]] = kv[1]
	return sh.report(nil)
}

func (sh *shell) git(args ...string) (string, int) {
	env := make([]string, len(sh.env))
	i := 0
	for n, v := range sh.env {
		env[i] = n + "=" + v
		i++
	}
	rc := 0
	cmd := exec.Command("git", args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		rc = 1
	}
	return string(out), rc
}

func (sh *shell) ln(args ...string) (string, int) {
	if len(args) != 3 || args[0] != "-s" {
		return sh.report(fmt.Errorf("invalid arguments"))
	}
	return sh.report(nazuna.CreateLink(args[1], args[2]))
}

func (sh *shell) ls(args ...string) (string, int) {
	f, err := os.Open(args[0])
	if err != nil {
		return sh.report(err)
	}
	defer f.Close()
	list, err := f.Readdir(-1)
	if err != nil {
		return sh.report(err)
	}
	rc := 0
	b := new(bytes.Buffer)
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
		fmt.Fprintf(b, "%v%v\n", fi.Name(), s)
	}
	return b.String(), rc
}

func (sh *shell) mkdir(args ...string) (string, int) {
	return sh.report(os.MkdirAll(args[0], 0777))
}

func (sh *shell) nzn(args ...string) (string, int) {
	reset := func(fs *cli.FlagSet) {
		fs.VisitAll(func(f *cli.Flag) {
			fs.Set(f.Name[0], f.Default)
		})
	}
	reset(app.Flags)
	for _, cmd := range app.Cmds {
		if cmd.Flags != nil {
			reset(cmd.Flags)
		}
	}

	b := new(bytes.Buffer)
	app.Stdout = b
	app.Stderr = b

	rc := 0
	if err := app.Run(args); err != nil {
		switch err := err.(type) {
		case cli.FlagError:
			rc = 2
		case SystemExit:
			rc = int(err)
		default:
			rc = 1
		}
	}
	return b.String(), rc
}

func (sh *shell) rm(args ...string) (string, int) {
	var remove func(string) error
	switch {
	case 1 < len(args) && args[0] == "-r":
		remove = nazuna.RemoveAll
		args = args[1:]
	default:
		remove = os.Remove
	}
	return sh.report(remove(args[0]))
}

func (sh *shell) setup(args ...string) (string, int) {
	for _, d := range []string{"h", "r", "w"} {
		out, rc := sh.mkdir(d)
		if rc != 0 {
			return out, rc
		}
	}
	out, rc := sh.export(sh.expand("HOME=$tempdir/h")...)
	if rc != 0 {
		return out, rc
	}
	for n, v := range sh.gitconfig {
		out, rc = sh.git("config", "--global", n, v)
		if rc != 0 {
			return out, rc
		}
	}
	return sh.report(nil)
}

func (sh *shell) touch(args ...string) (string, int) {
	return sh.report(ioutil.WriteFile(filepath.Clean(args[0]), []byte{}, 0666))
}

type script []*cmdLine

func (s script) exec() error {
	sh, err := newShell()
	if err != nil {
		return err
	}
	defer sh.exit()
	return sh.run(s)
}

type cmdLine struct {
	cmd []string
	out string
}

type lines struct {
	a []string
	b []string
}

func (d *lines) Equal(i, j int) bool {
	if strings.HasSuffix(d.a[i], " (re)") {
		m, err := regexp.MatchString(d.a[i][:len(d.a[i])-5], d.b[j])
		return err == nil && m
	}
	return d.a[i] == d.b[j]
}

func quote(path string) string {
	return regexp.QuoteMeta(filepath.FromSlash(path))
}
