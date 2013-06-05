//
// nazuna :: cli.go
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

package nazuna

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type CLI struct {
	Flag flag.FlagSet

	args []string
	in   io.Reader
	out  io.Writer
	err  io.Writer
	vars struct {
		help    bool
		version bool
	}
}

func NewCLI(args []string) *CLI {
	c := &CLI{
		args: args,
		in:   os.Stdin,
		out:  os.Stdout,
		err:  os.Stderr,
	}
	c.Flag.Init(c.args[0], flag.ContinueOnError)
	c.Flag.SetOutput(ioutil.Discard)
	c.Flag.BoolVar(&c.vars.help, "h", false, "")
	c.Flag.BoolVar(&c.vars.help, "help", false, "")
	c.Flag.BoolVar(&c.vars.version, "version", false, "")
	return c
}

func (c *CLI) Args() []string {
	return c.args
}

func (c *CLI) SetIn(in io.Reader) {
	c.in = in
}

func (c *CLI) SetOut(out io.Writer) {
	c.out = out
}

func (c *CLI) SetErr(err io.Writer) {
	c.err = err
}

func (c *CLI) Print(a ...interface{}) (int, error) {
	return fmt.Fprint(c.out, a...)
}

func (c *CLI) Printf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(c.out, format, a...)
}

func (c *CLI) Println(a ...interface{}) (int, error) {
	return fmt.Fprintln(c.out, a...)
}

func (c *CLI) Error(a ...interface{}) (int, error) {
	return fmt.Fprint(c.err, a...)
}

func (c *CLI) Errorf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(c.err, format, a...)
}

func (c *CLI) Errorln(a ...interface{}) (int, error) {
	return fmt.Fprintln(c.err, a...)
}

func (c *CLI) Run() int {
	if err := c.Flag.Parse(c.args[1:]); err != nil {
		return c.usage(2, nil, err)
	}
	var args []string
	switch {
	case c.vars.help:
		args = append(args, "help")
	case c.vars.version:
		args = append(args, "version")
	default:
		args = c.Flag.Args()
		if len(args) == 0 {
			return c.usage(1, nil, nil)
		}
	}

	cmd, err := FindCommand(Commands, args[0])
	if err != nil {
		return c.usage(1, nil, err)
	}
	if cmd.CustomFlags {
		args = args[1:]
	} else {
		cmd.Flag.Init(c.args[0], flag.ContinueOnError)
		cmd.Flag.SetOutput(ioutil.Discard)
		if err := cmd.Flag.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return c.usage(0, cmd, nil)
			}
			return c.usage(2, cmd, err)
		}
		args = cmd.Flag.Args()
	}
	if err := cmd.Run(c, args); err != nil {
		switch v := err.(type) {
		case *CommandError:
			return c.usage(1, nil, err)
		case FlagError:
			return c.usage(2, cmd, err)
		case SystemExit:
			return int(v)
		}
		c.Errorf("%s: %s\n", c.args[0], err)
		return 1
	}
	return 0
}

func (c *CLI) usage(rc int, cmd *Command, err error) int {
	if err != nil {
		if cmd != nil {
			c.Errorf("%s %s: %s\n", c.args[0], cmd.Name(), err)
		} else {
			c.Errorf("%s: %s\n", c.args[0], err)
		}
	}
	var args []string
	if cmd != nil {
		args = append(args, cmd.Name())
	}
	cmdHelp.Run(c, args)
	return rc
}

func (c *CLI) Exec(cmd *exec.Cmd) error {
	cmd.Stdin = c.in
	cmd.Stdout = c.out
	cmd.Stderr = c.err
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", cmd.Args[0], err)
	}
	return nil
}
