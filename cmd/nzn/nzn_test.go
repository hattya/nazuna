//
// nazuna/cmd/nzn :: nzn_test.go
//
//   Copyright (c) 2013-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hattya/go.cli"
	"github.com/hattya/go.diff"
	"github.com/hattya/nazuna"
)

func init() {
	nazuna.Discover(false)

	app.Name = "nzn"
}

func TestSystemExit(t *testing.T) {
	err := SystemExit(1)
	if g, e := err.Error(), "exit status 1"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

type shell struct {
	t         *testing.T
	dir       string
	env       map[string]string
	gitconfig map[string]string
	funcs     map[string]interface{}
}

func newShell(t *testing.T) (*shell, error) {
	dir := t.TempDir()
	sh := &shell{
		t:   t,
		dir: dir,
		env: map[string]string{
			"tempdir": dir,
		},
		gitconfig: map[string]string{
			"core.autocrlf": "false",
			"pull.rebase":   "false",
			"user.name":     "Nazuna",
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
	sh.t.Setenv("PWD", sh.dir)
	defer os.Chdir(wd)

	for i, c := range s {
		args := sh.expand(c.cmd[1:]...)
		out, rc := "unknown command", 1
		switch f := sh.funcs[c.cmd[0]].(type) {
		case func(*shell, ...string) (string, int):
			out, rc = f(sh, args...)
		case func(...string) (string, int):
			out, rc = f(args...)
		}
		if diff := sh.verify(c.out, out, rc); diff != "" {
			return fmt.Errorf("script:%d:\n$ %v\n%v", i+1, strings.Join(c.cmd, " "), diff)
		}
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
	var Δ []string
	format := func(sign string, lines []string, i, j int) {
		for ; i < j; i++ {
			Δ = append(Δ, sign+lines[i])
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
		if len(cl) > 0 {
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
	return strings.Join(Δ, "\n")
}

func (sh *shell) cat(args ...string) (string, int) {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return sh.report(err)
	}
	return string(data), 0
}

func (sh *shell) cd(args ...string) (string, int) {
	sh.t.Setenv("PWD", args[0])
	return sh.report(os.Chdir(args[0]))
}

func (sh *shell) export(args ...string) (string, int) {
	kv := strings.SplitN(args[0], "=", 2)
	sh.t.Setenv(kv[0], kv[1])
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
	list, err := os.ReadDir(args[0])
	if err != nil {
		return sh.report(err)
	}
	rc := 0
	var b strings.Builder
	for _, de := range list {
		var s string
		switch {
		case de.Type().IsRegular():
		case de.Type().IsDir():
			s = "/"
		case de.Type()&os.ModeSymlink != 0:
		default:
			s = ">"
			rc = 1
		}
		fmt.Fprintf(&b, "%v%v\n", de.Name(), s)
	}
	return b.String(), rc
}

func (sh *shell) mkdir(args ...string) (string, int) {
	return sh.report(os.MkdirAll(args[0], 0o777))
}

func (sh *shell) nzn(args ...string) (string, int) {
	app.Flags.Reset()
	for _, cmd := range app.Cmds {
		if cmd.Flags != nil {
			cmd.Flags.Reset()
		}
	}

	var b bytes.Buffer
	app.Stdout = &b
	app.Stderr = &b

	rc := 0
	if err := app.Run(args); err != nil {
		switch err := err.(type) {
		case cli.FlagError:
			rc = 2
		case cli.Interrupt:
			rc = 128 + 2
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
	if len(args) > 1 && args[0] == "-r" {
		remove = os.RemoveAll
		args = args[1:]
	} else {
		remove = os.Remove
	}
	return sh.report(remove(args[0]))
}

func (sh *shell) setup(args ...string) (string, int) {
	for _, d := range []string{"home", "public", "wc"} {
		out, rc := sh.mkdir(d)
		if rc != 0 {
			return out, rc
		}
		k := d
		if k == "home" {
			k = "HOME"
		}
		out, rc = sh.export(sh.expand(fmt.Sprintf("%v=$tempdir/%v", k, d))...)
		if rc != 0 {
			return out, rc
		}
	}
	for n, v := range sh.gitconfig {
		out, rc := sh.git("config", "--global", n, v)
		if rc != 0 {
			return out, rc
		}
	}
	return sh.report(nil)
}

func (sh *shell) touch(args ...string) (string, int) {
	return sh.report(os.WriteFile(filepath.Clean(args[0]), []byte{}, 0o666))
}

type script []*cmdLine

func (s script) exec(t *testing.T) error {
	sh, err := newShell(t)
	if err != nil {
		return err
	}
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

func path(path string) string {
	return filepath.FromSlash(path)
}

func quote(path string) string {
	return regexp.QuoteMeta(filepath.FromSlash(path))
}
