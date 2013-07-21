// Niels Widger
// Time-stamp: <21 Jul 2013 at 08:44:16 by nwidger on macros.local>

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

type Word string
type Value string

// FRAME

type Frame struct {
	level int
	vars map[string]Value
}

func NewFrame() *Frame {
	return &Frame{ level: 0, vars: make(map[string]Value) }
}

func (frame *Frame) GetValue(varName string) (ok bool, value Value) {
	value, ok = frame.vars[varName]
	return
}

func (frame *Frame) SetValue(varName string, value Value) (ok bool) {
	frame.vars[varName] = value
	return true
}

func (frame *Frame) GetLevel() int {
	return frame.level
}

func (frame *Frame) GetParentLevel() int {
	return frame.level - 1
}

// STACK

type Stack struct {
	level_list *list.List
	level_map map[int]*Frame
}

func (stack *Stack) PushFrame(frame *Frame) {
	top := 0

	if stack.level_list.Len() != 0 {
		top = stack.level_list.Front().Value.(*Frame).level
	}

	frame.level = top+1
	stack.level_map[frame.level] = frame
	stack.level_list.PushFront(frame)
}

func (stack *Stack) GetFrame(level int) (frame *Frame, error error) {
	frame, ok := stack.level_map[level]

	if !ok {
		return nil, errors.New("No frame at that level")
	}

	return frame, nil
}

func (stack *Stack) PeekFrame() *Frame {
	if stack.level_list.Len() != 0 {
		front := stack.level_list.Front()
		frame := front.Value.(*Frame)
		return frame
	}

	return nil
}

func (stack *Stack) PopFrame() {
	if stack.level_list.Len() != 0 {
		front := stack.level_list.Front()
		frame := front.Value.(*Frame)
		stack.level_map[frame.level] = nil
		stack.level_list.Remove(front)
	}
}

// NAMESPACE

type Namespace struct {
	name string
	commands map[string]Command
}

func NewNamespace(name string) Namespace {
	return Namespace{ name: name, commands: make(map[string]Command) }
}

func (ns *Namespace) AddCommand(cmd Command) bool {
	ns.commands[cmd.name] = cmd
	return true
}

func (ns *Namespace) FindCommand(name string, words *list.List) (cmd Command, error error) {
	var ok bool

	cmd, ok = ns.commands[name]; if !ok {
		error = errors.New("invalid command name \"%s\"\n")
		return
	}

	ok, error = cmd.ValidateArgs(words); if !ok {
		return
	}

	return cmd, error
}

// COMMAND

type Command struct {
	name string
	num_args int
	args []string

	body string
	native_body (func (*Interp, []string) string)
}

func (cmd *Command) BadArgsMessage() string {
	var err = fmt.Sprintf("wrong # args: should be \"%v", cmd.name)
	for _, arg := range cmd.args { err += fmt.Sprintf(" %v", arg) }
	err += "\""
	return err
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

// INTERP

type Interp struct {
	stack Stack
	namespaces map[string]Namespace

}

func NewInterp() *Interp {
	interp := &Interp{ Stack{ level_map: make(map[int]*Frame), level_list: list.New() }, make(map[string]Namespace) }
	interp.namespaces["::"] = NewNamespace("::")
	interp.AddBuiltinCommands()
	return interp
}

func (interp *Interp) AddCommand(ns_name string, cmd Command) bool {
	ns, ok := interp.namespaces[ns_name]

	if !ok {
		fmt.Println("no such namespace", ns_name)
		os.Exit(1)
	}

	return ns.AddCommand(cmd)
}

func (interp *Interp) AddBuiltinCommands() {
	ns_name := "::"

	interp.AddCommand(ns_name, Command{ "cd", 1, []string{ "dir" }, "", func (interp *Interp, args []string) string {
		os.Chdir(args[0])
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "eval", 1, []string{ "script" }, "", func (interp *Interp, args []string) string {
		return interp.Eval(args[1])
	}})

	interp.AddCommand(ns_name, Command{ "global", 1, []string{ "args" }, "", func (interp *Interp, args []string) string {
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "proc", 3, []string{ "name", "args", "body" }, "", func (interp *Interp, args []string) string {
		fields := strings.Fields(args[1])
		interp.AddCommand(ns_name, Command{ args[0], len(fields), fields, args[2], nil })
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "puts", 1, []string{ "string" }, "", func (interp *Interp, args []string) string {
		fmt.Println(args[0])
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "pwd", 0, []string{ "" }, "", func (interp *Interp, args []string) string {
		cwd, _ := os.Getwd()
		return cwd
	}})

	interp.AddCommand(ns_name, Command{ "set", -1, []string{ "varName", "value" }, "", func (interp *Interp, args []string) string {
		var ok bool
		var value Value

		varName := args[0]
		level := interp.stack.PeekFrame().GetParentLevel()
		frame, error := interp.stack.GetFrame(level)

		if error != nil {
			fmt.Println(error)
			os.Exit(1)
		}

		if len(args) == 2 {
			value = Value(args[1])
			frame.SetValue(varName, value)
		} else {
			ok, value = frame.GetValue(varName)

			if !ok {
				fmt.Printf("can't read \"%s\": no such variable\n", varName)
				os.Exit(1)
			}
		}

		return string(value)
	}})
}

func (interp *Interp) FindCommand(name string, words *list.List) (cmd Command, error error) {
	ns := interp.namespaces["::"]

	cmd, error = ns.FindCommand(name, words); if error != nil {
		return
	}

	ok, error := cmd.ValidateArgs(words); if !ok {
		return
	}

	return
}

func (interp *Interp) ParseDoubleQuoteWord(script string) (ok bool, word string, remainder string) {
	script = interp.SkipWhiteSpace(script)

	ok = false
	remainder = script

	loc := regexp.MustCompile("^\"([^\"])*\"").FindStringIndex(script); if loc == nil {
		return
	}

	ok = true
	word = script[loc[0]+1:loc[1]-1]
	remainder = interp.SkipWhiteSpace(string(script[loc[1]:]))

	return
}

func (interp *Interp) ParseBraceWord(script string) (ok bool, word string, remainder string) {
	script = interp.SkipWhiteSpace(script)

	ok = false
	remainder = script

	loc := regexp.MustCompile("^{([^}])*}").FindStringIndex(script); if loc == nil {
		return
	}

	ok = true
	word = script[loc[0]+1:loc[1]-1]
	remainder = interp.SkipWhiteSpace(string(script[loc[1]:]))

	return
}

func (interp *Interp) ParseWord(script string) (ok bool, word string, remainder string) {
	script = interp.SkipWhiteSpace(script)

	ok, word, remainder = interp.ParseDoubleQuoteWord(script); if ok {
		return
	}

	ok, word, remainder = interp.ParseBraceWord(script); if ok {
		return
	}

	ok = false
	remainder = script

	loc := regexp.MustCompile("^[^ \t\v\f\r\n;]+").FindStringIndex(script); if loc == nil {
		return
	}

	ok = true
	word = script[loc[0]:loc[1]]
	remainder = interp.SkipWhiteSpace(string(script[loc[1]:]))

	return
}

func (interp *Interp) ParseWords(script string) (ok bool, words *list.List, remainder string) {
	var word string

	remainder = script
	words = list.New()

	for len(remainder) != 0 {
		if string(remainder[0]) == "\n" || string(remainder[0]) == ";" {
			break
		}

		ok, word, remainder = interp.ParseWord(remainder); if !ok {
			fmt.Printf("error parsing next word from script \"%s\"\n", remainder)
			os.Exit(1)
		}

		words.PushBack(word)
	}

	return true, words, remainder
}

func (interp *Interp) SkipWhiteSpace(script string) (remainder string) {
	remainder = script

	loc := regexp.MustCompile("^[ \t\v\r\f]+").FindStringIndex(script); if loc == nil {
		return
	}

	remainder = script[loc[1]:]
	return
}

func (interp *Interp) wordListToArray(words *list.List) []string {
	i := 0
	swords := make([]string, words.Len())

	for e := words.Front(); e != nil; e = e.Next() {
		swords[i] = e.Value.(string)
		i++
	}

	return swords
}

func (interp *Interp) Eval(script string) string {
	var words *list.List
	var error error
	var cmd Command

	retval := ""
	interp.stack.PushFrame(NewFrame())

	for len(script) != 0 {
		_, words, script = interp.ParseWords(script)

		if len(script) > 0 && (string(script[0]) == "\n" || string(script[0]) == ";") {
			script = script[1:]

			if words.Len() == 0 {
				continue
			}

			name := words.Remove(words.Front()).(string)
			cmd, error = interp.FindCommand(name, words)

			if error != nil {
				fmt.Println(error)
				os.Exit(1)
			}

			// TODO: bind arguments to stack level

			if cmd.native_body == nil {
				retval = interp.Eval(cmd.body)
			} else {
				interp.stack.PushFrame(NewFrame())
				retval = cmd.native_body(interp, interp.wordListToArray(words))
				interp.stack.PopFrame()
			}

		}
	}

	interp.stack.PopFrame()

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
	interp := NewInterp()
	interp.Eval(script)
}
