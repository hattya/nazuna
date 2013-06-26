//
// nazuna :: command.go
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
	"sort"
	"strings"
)

type Command struct {
	Names       []string
	Usage       string
	Help        string
	Flag        flag.FlagSet
	CustomFlags bool
	Run         func(UI, []string) error
}

func (c *Command) Name() string {
	if len(c.Names) == 0 {
		return ""
	}
	return c.Names[0]
}

type FlagError string

func (e FlagError) Error() string {
	return string(e)
}

var Commands = []*Command{
	cmdAlias,
	cmdClone,
	cmdHelp,
	cmdInit,
	cmdLayer,
	cmdLink,
	cmdSubrepo,
	cmdUpdate,
	cmdVCS,
	cmdVersion,
}

type CommandError struct {
	Name string
	List []string
}

func (e *CommandError) Error() string {
	if len(e.List) == 0 {
		return fmt.Sprintf("unknown command '%s'", e.Name)
	}
	return fmt.Sprintf("command '%s' is ambiguous:\n    %s", e.Name, strings.Join(e.List, " "))
}

func FindCommand(commands []*Command, name string) (cmd *Command, err error) {
	set := make(map[string]*Command)
loop:
	for _, c := range commands {
		if c.Run != nil {
			for _, n := range c.Names {
				if n == name {
					set[n] = c
					continue loop
				}
			}
			for _, n := range c.Names {
				if strings.HasPrefix(n, name) {
					set[n] = c
					continue loop
				}
			}
		}
	}

	switch len(set) {
	case 0:
		err = &CommandError{Name: name}
	case 1:
		for _, cmd = range set {
		}
	default:
		if c, found := set[name]; found {
			cmd = c
		} else {
			list := make([]string, len(set))
			i := 0
			for n, _ := range set {
				list[i] = n
				i++
			}
			err = &CommandError{name, list}
		}
	}
	return
}

func sortCommands(commands []*Command) []*Command {
	list := make(commandByName, len(commands))
	for i, c := range commands {
		list[i] = c
	}
	sort.Sort(list)
	return list
}

type commandByName []*Command

func (s commandByName) Len() int           { return len(s) }
func (s commandByName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
func (s commandByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
