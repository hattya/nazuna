//
// nazuna :: help.go
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
	"strings"
	"unicode"
)

var cmdHelp = &Command{
	Names: []string{"help"},
	Usage: "help [options] [--] [command]",
	Help: `
  display help information about nazuna
`,
}

func init() {
	cmdHelp.Run = runHelp
}

func runHelp(ui UI, args []string) (err error) {
	var cmd *Command
	if 0 < len(args) {
		cmd, err = FindCommand(Commands, args[0])
		if err != nil {
			return
		}
	}

	if cmd == nil {
		ui.Print("nazuna - A layered dotfiles management\n\n")
		ui.Print("list of commands:\n\n")
		maxWidth := 0
		for _, cmd := range Commands {
			if w := len(cmd.Name()); maxWidth < w {
				maxWidth = w
			}
		}
		for _, cmd := range sortCommands(Commands) {
			l := strings.SplitN(strings.TrimSpace(cmd.Help), "\n", 2)[0]
			ui.Printf("  %-*s    %s\n", maxWidth, cmd.Name(), l)
		}
		ui.Println()
	} else {
		ui.Printf("usage: %s %s\n", ui.Args()[0], cmd.Usage)
		for _, l := range strings.Split(cmd.Help, "\n") {
			ui.Println(strings.TrimRightFunc(l, unicode.IsSpace))
		}
	}
	return
}
