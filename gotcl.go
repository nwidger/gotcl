// Niels Widger
// Time-stamp: <22 Jul 2013 at 21:00:23 by nwidger on macros.local>

package main

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

// WORD

type Word string

// VALUE

type Value string

func (value Value) String() string {
	return string(value)
}

// FRAME

type Frame struct {
	level int
	vars map[string]Value
}

func NewFrame() *Frame {
	return &Frame{ level: 0, vars: make(map[string]Value) }
}

func (frame *Frame) GetValue(varName string) (value Value, ok bool) {
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

func (frame *Frame) BindArguments(cmd Command, words *list.List) bool {
	for _, arg := range cmd.args {
		name := arg.GetName()
		value := arg.GetValue()
		// fmt.Printf("111 %v: name ->%v<-, value ->%v<-\n", cmd.name, name, value)
		if words.Len() != 0 { value = words.Remove(words.Front()).(string) }
		// fmt.Printf("222 %v: name ->%v<-, value ->%v<-\n", cmd.name, name, value)
		frame.SetValue(name, Value(value))
	}

	return true
}

// STACK

type Stack struct {
	level_list *list.List
	level_map map[int]*Frame
}

func (stack *Stack) PushFrame() *Frame {
	frame := NewFrame()
	top := 0

	if stack.level_list.Len() != 0 {
		top = stack.level_list.Front().Value.(*Frame).level
	}

	frame.level = top+1
	stack.level_map[frame.level] = frame
	stack.level_list.PushFront(frame)

	return frame
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
	vars map[string]Value
	commands map[string]Command
}

func NewNamespace(name string) Namespace {
	return Namespace{ name: name, vars: make(map[string]Value), commands: make(map[string]Command) }
}

func (namespace *Namespace) GetValue(varName string) (value Value, ok bool) {
	value, ok = namespace.vars[varName]
	return
}

func (namespace *Namespace) SetValue(varName string, value Value) (ok bool) {
	namespace.vars[varName] = value
	return true
}

func (ns *Namespace) AddCommand(cmd Command) bool {
	ns.commands[cmd.name] = cmd
	return true
}

func (ns *Namespace) FindCommand(name string, words *list.List) (cmd Command, error error) {
	var ok bool

	cmd, ok = ns.commands[name]; if !ok {
		msg := fmt.Sprintf("invalid command name \"%s\"\n", name)
		error = errors.New(msg)
		return
	}

	ok, error = cmd.ValidateArgs(words); if !ok {
		return
	}

	return cmd, error
}

// ARG

type Arg struct {
	name string
	has_default_value bool
	value string
}

func NewArg(name string) Arg {
	return Arg{ name: name, has_default_value: false, value: "" }
}

func NewArgDefault(name string, value string) Arg {
	return Arg{ name: name, has_default_value: true, value: value }
}

func (arg *Arg) HasDefaultValue() bool {
	return arg.has_default_value
}

func (arg *Arg) GetName() string {
	return arg.name
}

func (arg *Arg) GetValue() string {
	return arg.value
}

func (arg *Arg) String() string {
	str := fmt.Sprintf("[ name: ->%v<-\n", arg.name)
	str += fmt.Sprintf("has_default_value: ->%v<-\n", arg.has_default_value)
	str += fmt.Sprintf("value: ->%v<- ]\n", arg.value)
	return str
}

// COMMAND

type Command struct {
	name string
	num_args int
	min_args int
	args []Arg

	body string
	native_body (func (*Interp) string)
}

func (cmd *Command) BadArgsMessage() string {
	var err = fmt.Sprintf("wrong # args: should be \"%v", cmd.name)

	for _, arg := range cmd.args {
		err += " "
		if arg.HasDefaultValue() { err += "?" }
		err += fmt.Sprintf("%v", arg.name)
		if arg.HasDefaultValue() { err += "?" }
	}

	err += "\""

	return err
}

func (cmd *Command) ValidateArgs(words *list.List) (ok bool, err error) {
	ok = true
	err = nil

	if words.Len() < cmd.min_args || words.Len() > cmd.num_args  {
		ok = false
		err = errors.New(cmd.BadArgsMessage())
	}

	return
}

func (cmd *Command) String() string {
	str := fmt.Sprintf("name: ->%v<-\n", cmd.name)
	str += fmt.Sprintf("num_args: ->%v<-\n", cmd.num_args)
	str += fmt.Sprintf("min_args: ->%v<-\n", cmd.min_args)
	for i, arg := range cmd.args {
		str += fmt.Sprintf("arg[%v]: ->%v<-\n", i, arg.String())
	}
	str += fmt.Sprintf("body: ->%v<-\n", cmd.body)
	return str
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
	interp.stack.PushFrame()
	return interp
}

func (interp *Interp) AddCommand(ns_name string, cmd Command) bool {
	ns, ok := interp.namespaces[ns_name]

	if !ok {
		fmt.Printf("can't create procedure \"%v\": unknown namespace\n", cmd.name)
		os.Exit(1)
	}

	// fmt.Println("adding command", cmd.String())

	return ns.AddCommand(cmd)
}

func (interp *Interp) AddBuiltinCommands() {
	ns_name := "::"

	interp.AddCommand(ns_name, Command{ "cd", 1, 1, []Arg{ NewArg("dir") }, "", func (interp *Interp) string {
		frame := interp.stack.PeekFrame()
		dir, _ := frame.GetValue("dir")
		os.Chdir(string(dir))
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "eval", 1, 1, []Arg{ NewArg("script") }, "", func (interp *Interp) string {
		frame := interp.stack.PeekFrame()
		script, _ := frame.GetValue("script")
		return interp.Eval(string(script))
	}})

	interp.AddCommand(ns_name, Command{ "global", 1, 1, []Arg{ NewArg("args") }, "", func (interp *Interp) string {
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "if", 1, 1, []Arg{ NewArg("args") }, "", func (interp *Interp) string {
		// frame := interp.stack.PeekFrame()
		// args, _ := frame.GetValue("args")
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "proc", 3, 3, []Arg{ NewArg("name"), NewArg("args"), NewArg("body") }, "", func (interp *Interp) string {
		frame := interp.stack.PeekFrame()

		name, _ := frame.GetValue("name")
		args, _ := frame.GetValue("args")
		body, _ := frame.GetValue("body")

		_, words, _ := interp.ParseWords(args.String())

		num_args := words.Len()
		min_args := num_args

		s_args := make([]Arg, num_args)

		i := 0
		for e := words.Front(); e != nil; e = e.Next() {
			_, arg_words, _ := interp.ParseWords(e.Value.(string))
			arg_name := arg_words.Remove(arg_words.Front()).(string)

			if arg_words.Len() == 0 {
				s_args[i] = NewArg(arg_name)
			} else {
				arg_value := arg_words.Remove(arg_words.Front()).(string)
				s_args[i] = NewArgDefault(arg_name, arg_value)
				if min_args == num_args { min_args = i }
			}

			i++
		}

		interp.AddCommand(ns_name, Command{ name.String(), num_args, min_args, s_args, body.String(), nil })
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "puts", 1, 1, []Arg{ NewArg("string") }, "", func (interp *Interp) string {
		frame := interp.stack.PeekFrame()
		str, _ := frame.GetValue("string")
		fmt.Println(string(str))
		return ""
	}})

	interp.AddCommand(ns_name, Command{ "pwd", 0, 0, []Arg{ }, "", func (interp *Interp) string {
		cwd, _ := os.Getwd()
		return cwd
	}})

	interp.AddCommand(ns_name, Command{ "set", 2, 1, []Arg{ NewArg("varName"), NewArgDefault("newValue", "") }, "", func (interp *Interp) string {
		var ok bool
		var value Value

		frame := interp.stack.PeekFrame()
		varName, _ := frame.GetValue("varName")
		value, _ = frame.GetValue("newValue")

		// fmt.Printf("in set, varName = ->%v<-, value = ->%v<-\n", varName, value)

		level := frame.GetParentLevel()
		frame, error := interp.stack.GetFrame(level)

		if error != nil {
			fmt.Println(error)
			os.Exit(1)
		}

		if string(value) != "" {
			frame.SetValue(string(varName), value)
		} else {
			value, ok = frame.GetValue(string(varName))

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

	// loc := regexp.MustCompile("^{([^}])*}").FindStringIndex(script); if loc == nil {
	loc := regexp.MustCompile("^{(({[^}]*})|[^}])*}").FindStringIndex(script); if loc == nil {
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
		if remainder[0] == '\n' || remainder[0] == ';' {
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

func (interp *Interp) ParseComment(script string) (ok bool, comment string, remainder string) {
	script = interp.SkipWhiteSpace(script)

	ok = false
	comment = ""
	remainder = script

	loc := regexp.MustCompile("^#[^\n]*\n").FindStringIndex(script); if loc == nil {
		return
	}

	ok = true
	comment = script[loc[0]:loc[1]]
	remainder = script[loc[1]:]

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
	var error error
	var cmd Command
	var words *list.List

	retval := ""

	for len(script) != 0 {
		ok, _, script = interp.ParseComment(script); if ok {
			continue
		}

		_, words, script = interp.ParseWords(script)

		if len(script) > 0 && (script[0] == '\n' || script[0] == ';') {
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

			frame := interp.stack.PushFrame()
			frame.BindArguments(cmd, words)

			if cmd.native_body == nil {
				retval = interp.Eval(cmd.body)
			} else {
				retval = cmd.native_body(interp)
			}

			interp.stack.PopFrame()
		}
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
	interp := NewInterp()
	interp.Eval(script)
}
