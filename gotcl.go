// Niels Widger
// Time-stamp: <17 Jul 2013 at 20:04:48 by nwidger on macros.local>

package main

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// SCRIPT = COMMANDS
//
// COMMANDS = COMMANDS COMMAND CMD_SEP
//          | COMMAND CMD_SEP
//
// COMMAND = WORDS
//
// WORDS = WORDS WORD
//       | WORD
//
// WORD = " DQUOTE_WORD_CHARS "
//      | { BRACE_WORD_CHARS }
//      | {*} ARG_SUB_WORD_CHARS
//      | WORD_CHARS

var COMMAND_SEP regexp.Regexp = *regexp.MustCompile("[;\n]")

var WORD_SEP regexp.Regexp = *regexp.MustCompile("[ \t\f]")

var ARG_EXP regexp.Regexp = *regexp.MustCompile("{\\*}[^ \t\f]")

var NAME regexp.Regexp = *regexp.MustCompile("([a-zA-Z0-9_]|::)+")

// var BACKSLASH_SUB regexp.Regexp = *regexp.MustCompile("\\([abfnrtv\\]|[0-9]{1-3}|(x([0-9a-fA-F])*)|(u([0-9a-fA-F]){1-4}))")

type Word struct { }

type Command struct {
	name string
	num_args int
	args []string

	body string
	native_body (func (*Interp, []string) string)
}

type Interp struct {
	commands map[string]Command

}

func MakeInterp() Interp {
	interp := Interp{ make(map[string]Command) }
	interp.AddBuiltinCommands()
	return interp
}

func (interp *Interp) AddCommand(cmd Command) bool {
	interp.commands[cmd.name] = cmd
	return true
}

func (interp *Interp) AddBuiltinCommands() {
	interp.AddCommand(Command{ "cd", 1, []string{ "dir" }, "", func (interp *Interp, args []string) string {
		os.Chdir(args[0])
		return ""
	}})

	interp.AddCommand(Command{ "eval", 1, []string{ "script" }, "", func (interp *Interp, args []string) string {
		return interp.Eval(args[1])
	}})

	interp.AddCommand(Command{ "proc", 3, []string{ "name", "args", "body" }, "", func (interp *Interp, args []string) string {
		fields := strings.Fields(args[1])
		interp.AddCommand(Command{ args[0], len(fields), fields, args[2], nil })
		return ""
	}})

	interp.AddCommand(Command{ "puts", 1, []string{ "string" }, "", func (interp *Interp, args []string) string {
		fmt.Println(args[0])
		return ""
	}})

	interp.AddCommand(Command{ "pwd", 0, []string{ "" }, "", func (interp *Interp, args []string) string {
		cwd, _ := os.Getwd()
		return cwd
	}})
}

func (cmd *Command) ValidateArgs(words *list.List) (ok bool, err error) {
	ok = true
	err = nil

	if cmd.num_args != -1 && cmd.num_args != words.Len() {
		ok = false
		err = errors.New(cmd.BadArgsMessage())
	}

	return
}

func (cmd *Command) BadArgsMessage() string {
	var err = fmt.Sprintf("wrong # args: should be \"%v", cmd.name)
	for _, arg := range cmd.args { err += fmt.Sprintf(" %v", arg) }
	err += "\""
	return err
}

func (interp *Interp) ParseDoubleQuoteWord(script string) (ok bool, word string, remainder string) {
	ok = false
	remainder = script

	loc := regexp.MustCompile("^\"([^\"])*\"").FindStringIndex(script); if loc == nil {
		return
	}

	ok = true
	word = script[loc[0]+1:loc[1]-1]
	remainder = string(script[loc[1]:])

	return
}

func (interp *Interp) ParseBraceWord(script string) (ok bool, word string, remainder string) {
	ok = false
	remainder = script

	loc := regexp.MustCompile("^{([^}])*}").FindStringIndex(script); if loc == nil {
		return
	}

	ok = true
	word = script[loc[0]+1:loc[1]-1]
	remainder = string(script[loc[1]:])

	return
}

func (interp *Interp) ParseWord(script string) (ok bool, word string, remainder string) {
	ok, word, remainder = interp.ParseDoubleQuoteWord(script); if ok {
		return
	}

	ok, word, remainder = interp.ParseBraceWord(script); if ok {
		return
	}

	ok = false
	remainder = script

	loc := regexp.MustCompile("^[^ \t\v\r\f\n]+").FindStringIndex(script); if loc == nil {
		return
	}

	ok = true
	word = script[loc[0]:loc[1]]
	remainder = string(script[loc[1]:])

	return
}

func (interp *Interp) SkipWhiteSpace(script string) (remainder string) {
	remainder = script

	loc := regexp.MustCompile("^[ \t\v\r\f]+").FindStringIndex(script); if loc == nil {
		return
	}

	remainder = script[loc[1]:]
	return
}

func (interp *Interp) Eval(script string) string {
	var ok bool
	var word string

	retval := ""
	script = interp.SkipWhiteSpace(script)

	for len(script) != 0 {
		words := list.New()

		for len(script) != 0 {
			if (string(script[0]) == "\n" || string(script[0]) == ";") {
				script = script[1:]
				script = interp.SkipWhiteSpace(script)
				break
			}
			
			ok, word, script = interp.ParseWord(script); if !ok {
				fmt.Printf("error parsing next word from script \"%s\"\n", script)
				os.Exit(1)
			}

			words.PushBack(word)
			script = interp.SkipWhiteSpace(script)
		}

		if words.Len() == 0 {
			continue
		}

		name := words.Remove(words.Front()).(string)
		cmd, ok := interp.commands[name]; if !ok {
			fmt.Printf("invalid command name \"%s\"\n", name)
			os.Exit(1)
		}

		// TODO: bind arguments to stack level
		ok, error := cmd.ValidateArgs(words); if !ok {
			fmt.Println(error)
			os.Exit(1)
		}

		if cmd.native_body == nil {
			retval = interp.Eval(cmd.body)
		} else {
			i := 0
			swords := make([]string, words.Len())

			for e := words.Front(); e != nil; e = e.Next() {
				swords[i] = e.Value.(string)
				i++
			}

			retval = cmd.native_body(interp, swords)
		}

		script = interp.SkipWhiteSpace(script)
	}

	return retval
}

func usage() {
	fmt.Printf("%s: FILE\n", os.Args[0])
}

func main() {
	if (len(os.Args) != 2) {
		usage()
		os.Exit(1)
	}

	bytes, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		fmt.Println("Error opening file %s: %s\n", os.Args[1], err)
		os.Exit(1)
	}

	script := string(bytes)
	interp := MakeInterp()
	interp.Eval(script)
}
