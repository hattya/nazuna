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
	if g, e := len(args), 1; g != e {
		t.Fatalf("expected %v, got %v", e, g)
	}
	if g, e := args[0], "nazuna.test"; g != e {
		t.Errorf("expected %q, got %q", e, g)
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
	if g, e := out.String(), "Print()\nPrintf()\nPrintln()\n"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	c.Error("Error()\n")
	c.Errorf("%s\n", "Errorf()")
	c.Errorln("Errorln()")
	if g, e := err.String(), "Error()\nErrorf()\nErrorln()\n"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestProcess", "0")
	if err := c.Exec(cmd); err != nil {
		t.Error(err)
	}

	cmd = exec.Command(os.Args[0], "-test.run=TestProcess", "7")
	re := regexp.MustCompile(regexp.QuoteMeta(os.Args[0]) + `: exit status.* 7`)
	switch err := c.Exec(cmd); {
	case err == nil:
		t.Error("expected error")
	case !re.MatchString(err.Error()):
		t.Error("unexpected error:", err)
	}
}

func TestRunCLI(t *testing.T) {
	s := script{
		{
			cmd: []string{"nzn", "--nazuna"},
			out: fmt.Sprintf("nzn: flag .* not defined: -*nazuna (re)\n%s[2]\n", helpOut),
		},
		{
			cmd: []string{"nzn", "--help"},
			out: helpOut,
		},
		{
			cmd: []string{"nzn", "--version"},
			out: versionOut,
		},
		{
			cmd: []string{"nzn"},
			out: fmt.Sprintf("%s[1]\n", helpOut),
		},
		{
			cmd: []string{"nzn", "nazuna"},
			out: fmt.Sprintf("nzn: unknown command 'nazuna'\n%s[1]\n", helpOut),
		},
		{
			cmd: []string{"nzn", "help", "--help"},
			out: helpUsage,
		},
		{
			cmd: []string{"nzn", "help", "--nazuna"},
			out: fmt.Sprintf("nzn help: flag .* not defined: -*nazuna (re)\n%s[2]\n", helpUsage),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestProcess(*testing.T) {
	if len(os.Args) != 3 || os.Args[1] != "-test.run=TestProcess" {
		return
	}
	n, _ := strconv.Atoi(os.Args[2])
	os.Exit(n)
}
