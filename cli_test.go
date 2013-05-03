//
// nazuna :: cli_test.go
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
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"

	"github.com/hattya/nazuna"
)

func TestCLI(t *testing.T) {
	c := nazuna.NewCLI([]string{"nazuna.test"})
	args := c.Args()
	switch {
	case len(args) != 1:
		t.Errorf("expected 1, got %d", len(args))
	case args[0] != "nazuna.test":
		t.Errorf(`expected "nazuna.test", got %q`, args[0])
	}

	in := new(bytes.Buffer)
	out := new(bytes.Buffer)
	err := new(bytes.Buffer)
	c.SetIn(in)
	c.SetOut(out)
	c.SetErr(err)

	c.Print("Print()\n")
	c.Printf("%s\n", "Printf()")
	c.Println("Println()")
	c.Error("Error()\n")
	c.Errorf("%s\n", "Errorf()")
	c.Errorln("Errorln()")

	expected := `Print()
Printf()
Println()
`
	if err := equal(expected, out.String()); err != nil {
		t.Error(err)
	}

	expected = `Error()
Errorf()
Errorln()
`
	if err := equal(expected, err.String()); err != nil {
		t.Error(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestProcess", "0")
	if err := c.Exec(cmd); err != nil {
		t.Error(err)
	}

	cmd = exec.Command(os.Args[0], "-test.run=TestProcess", "7")
	re := regexp.MustCompile("nazuna.test: exit status.* 7")
	switch err := c.Exec(cmd); {
	case err == nil:
		t.Error("error expected")
	case !re.MatchString(err.Error()):
		t.Errorf("expected %q, got %q", re, err)
	}
}

func TestRunCLI(t *testing.T) {
	rc, bout, berr := runCLI("nazuna.test", "--nazuna")
	if rc != 2 {
		t.Errorf("expected 2, got %d", rc)
	}
	if err := equal(nazuna.HelpOut, bout); err != nil {
		t.Error(err)
	}
	re := regexp.MustCompile("^nazuna.test: flag .* not defined:")
	if !re.MatchString(berr) {
		t.Errorf("expected %q, got %q", re, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "--help")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if err := equal(nazuna.HelpOut, bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "--version")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if err := equal(nazuna.VersionOut, bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test")
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if err := equal(nazuna.HelpOut, bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "nazuna")
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if err := equal(nazuna.HelpOut, bout); err != nil {
		t.Error(err)
	}
	if berr != "nazuna.test: unknown command 'nazuna'\n" {
		t.Errorf(`expected "nazuna.test: unknown command 'nazuna'\n", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "help", "--help")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if err := equal(nazuna.HelpUsage, bout); err != nil {
		t.Error(err)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}

	rc, bout, berr = runCLI("nazuna.test", "help", "--nazuna")
	if rc != 2 {
		t.Errorf("expected 2, got %d", rc)
	}
	if err := equal(nazuna.HelpUsage, bout); err != nil {
		t.Error(err)
	}
	re = regexp.MustCompile("^nazuna.test help: flag .* not defined:")
	if !re.MatchString(berr) {
		t.Errorf("expected %q, got %q", re, berr)
	}
}

func TestProcess(*testing.T) {
	if len(os.Args) != 3 || os.Args[1] != "-test.run=TestProcess" {
		return
	}
	n, _ := strconv.Atoi(os.Args[2])
	os.Exit(n)
}

func runCLI(args ...string) (int, string, string) {
	for _, c := range nazuna.Commands {
		c.Flag.Visit(func(f *flag.Flag) {
			c.Flag.Set(f.Name, f.DefValue)
		})
	}

	out := new(bytes.Buffer)
	err := new(bytes.Buffer)
	c := nazuna.NewCLI(args)
	c.SetOut(out)
	c.SetErr(err)
	return c.Run(), out.String(), err.String()
}

func equal(expected, actual string) error {
	if actual == expected {
		return nil
	}
	return fmt.Errorf(`strings not equal:
<< expected >>
%s--
<<   got    >>
%s--
`, expected, actual)
}
