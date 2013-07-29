// Niels Widger
// Time-stamp: <29 Jul 2013 at 19:39:36 by nwidger on macros.local>

package main

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

// WORD

type Word interface {
	String() string
}

type LiteralWord string

func (word LiteralWord) String() string { return string(word) }

type BraceWord string

func (word BraceWord) String() string { return string(word) }

type DoubleQuoteWord string

func (word DoubleQuoteWord) String() string { return string(word) }

// VALUE

type Value string

func (value Value) Len() int {
	return len(string(value))
}

func (value Value) String() string {
	return string(value)
}

func (value Value) Int() int {
	i, err := strconv.Atoi(value.String())

	if err != nil {
		fmt.Printf("Cannot convert \"%s\" to an integer\n", value.String())
		os.Exit(1)
	}

	return i
}

func StringToValue(str string) Value {
	return Value(str)
}

func StringToValueP(str string) *Value {
	value := Value(str)
	return &value
}

func IntToValue(i int) Value {
	str := strconv.FormatInt(int64(i), 10)
	return Value(str)
}

func IntToValueP(i int) *Value {
	str := strconv.FormatInt(int64(i), 10)
	value := Value(str)
	return &value
}

// FRAME

type Frame struct {
	level int
	vars  map[string]*Value
}

func NewFrame() *Frame {
	return &Frame{level: 0, vars: make(map[string]*Value)}
}

func (frame *Frame) GetValue(varName string) (value *Value, ok bool) {
	value, ok = frame.vars[varName]
	return
}

func (frame *Frame) SetValue(varName string, value *Value) (ok bool) {
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

		if words.Len() != 0 {
			word := Value(words.Remove(words.Front()).(Word).String())
			value = &word
		}

		frame.SetValue(name, value)
	}

	return true
}

func (frame *Frame) SubstituteCommand(interp *Interp, word Word) (ok bool, newword Word, remainder Word) {
	ok = false
	newword = LiteralWord("")
	remainder = word

	loc := regexp.MustCompile("\\[((\\[[^\\]]*\\])|[^\\]])*\\]").FindStringIndex(word.String())
	if loc == nil {
		return
	}

	ok = true
	script := LiteralWord(word.String()[loc[0]+1 : loc[1]-1])
	newword = LiteralWord(newword.String() + word.String()[:loc[0]] + interp.Eval(script.String()))
	remainder = LiteralWord(word.String()[loc[1]:])

	return
}

func (frame *Frame) SubstituteVariable(interp *Interp, word Word) (ok bool, newword Word, remainder Word) {
	ok = false
	newword = LiteralWord("")
	remainder = word

	loc := regexp.MustCompile("\\$([a-zA-Z0-9_]|::)+").FindStringIndex(word.String())
	if loc == nil {
		return
	}

	ok = true
	name := word.String()[loc[0]+1 : loc[1]]

	value, ok := frame.GetValue(name)
	if !ok {
		fmt.Printf("can't read \"%s\": no such variable\n", name)
		os.Exit(1)
	}

	newword = LiteralWord(newword.String() + word.String()[:loc[0]] + value.String())
	remainder = LiteralWord(word.String()[loc[1]:])

	return
}

func (frame *Frame) Substitute(interp *Interp, word Word) Word {
	var ok bool
	var part Word

	switch word.(type) {
	case BraceWord:
		return word
	}

	newword := LiteralWord("")

	for len(word.String()) != 0 {
		ok, part, word = frame.SubstituteCommand(interp, word)
		if ok {
			newword = LiteralWord(newword.String() + part.String())
			continue
		}

		ok, part, word = frame.SubstituteVariable(interp, word)
		if ok {
			newword = LiteralWord(newword.String() + part.String())
			continue
		}

		newword = LiteralWord(newword.String() + word.String())
		word = LiteralWord("")
	}

	return newword
}

func (frame *Frame) SubstituteWords(interp *Interp, words *list.List) *list.List {
	newwords := list.New()

	for e := words.Front(); e != nil; e = e.Next() {
		newwords.PushBack(frame.Substitute(interp, e.Value.(Word)))
	}

	return newwords
}

func (frame *Frame) Eval(interp *Interp, script string) string {
	var ok bool
	var words *list.List

	retval := ""

	for len(script) != 0 {
		ok, _, script = ParseComment(script)
		if ok {
			continue
		}

		_, words, script = ParseWords(script)

		if len(script) == 0 || script[0] == '\n' || script[0] == ';' {
			if len(script) != 0 {
				script = script[1:]
			}

			if words.Len() == 0 {
				continue
			}

			retval = interp.EvalWords(words)
		}
	}

	return retval

}

// STACK

type Stack struct {
	level_list *list.List
	level_map  map[int]*Frame
}

func (stack *Stack) PushFrame() *Frame {
	frame := NewFrame()
	top := 0

	if stack.level_list.Len() != 0 {
		top = stack.level_list.Front().Value.(*Frame).level
	}

	frame.level = top + 1
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
	name     string
	vars     map[string]*Value
	commands map[string]Command
}

func NewNamespace(name string) Namespace {
	return Namespace{name: name, vars: make(map[string]*Value), commands: make(map[string]Command)}
}

func (namespace *Namespace) GetValue(varName string) (value *Value, ok bool) {
	value, ok = namespace.vars[varName]
	return
}

func (namespace *Namespace) SetValue(varName string, value *Value) (ok bool) {
	namespace.vars[varName] = value
	return true
}

func (ns *Namespace) AddCommand(cmd Command) bool {
	ns.commands[cmd.name] = cmd
	return true
}

func (ns *Namespace) FindCommand(name string, words *list.List) (cmd Command, error error) {
	var ok bool

	cmd, ok = ns.commands[name]
	if !ok {
		msg := fmt.Sprintf("invalid command name \"%s\"\n", name)
		error = errors.New(msg)
		return
	}

	ok, error = cmd.ValidateArgs(words)
	if !ok {
		return
	}

	return cmd, error
}

// ARG

type Arg struct {
	name              string
	has_default_value bool
	value             *Value
}

func NewArg(name string) Arg {
	return Arg{name: name, has_default_value: false, value: nil}
}

func NewArgDefault(name string, value *Value) Arg {
	return Arg{name: name, has_default_value: true, value: value}
}

func (arg *Arg) HasDefaultValue() bool {
	return arg.has_default_value
}

func (arg *Arg) GetName() string {
	return arg.name
}

func (arg *Arg) GetValue() *Value {
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
	name     string
	num_args int
	min_args int
	args     []Arg

	body        string
	native_body (func(*Interp, *Frame) string)
}

func (cmd *Command) BadArgsMessage() string {
	var err = fmt.Sprintf("wrong # args: should be \"%v", cmd.name)

	for _, arg := range cmd.args {
		err += " "
		if arg.HasDefaultValue() {
			err += "?"
		}
		err += fmt.Sprintf("%v", arg.name)
		if arg.HasDefaultValue() {
			err += "?"
		}
	}

	err += "\""

	return err
}

func (cmd *Command) ValidateArgs(words *list.List) (ok bool, err error) {
	ok = true
	err = nil

	if words.Len() < cmd.min_args || words.Len() > cmd.num_args {
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
	stack      Stack
	namespaces map[string]Namespace
}

func NewInterp() *Interp {
	interp := &Interp{Stack{level_map: make(map[int]*Frame), level_list: list.New()}, make(map[string]Namespace)}
	interp.namespaces["::"] = NewNamespace("::")
	interp.AddBuiltinCommands()
	return interp
}

func (interp *Interp) AddCommand(ns_name string, cmd Command) bool {
	ns, ok := interp.namespaces[ns_name]

	if !ok {
		fmt.Printf("can't create procedure \"%v\": unknown namespace\n", cmd.name)
		os.Exit(1)
	}

	return ns.AddCommand(cmd)
}

func (interp *Interp) AddBuiltinCommands() {
	ns_name := "::"

	interp.AddCommand(ns_name, Command{"cd", 1, 1, []Arg{NewArg("dir")}, "", func(interp *Interp, frame *Frame) string {
		dir, _ := frame.GetValue("dir")
		os.Chdir(dir.String())
		return ""
	}})

	interp.AddCommand(ns_name, Command{"eval", 1, 1, []Arg{NewArg("script")}, "", func(interp *Interp, frame *Frame) string {
		script, _ := frame.GetValue("script")
		return interp.Eval(script.String())
	}})

	interp.AddCommand(ns_name, Command{"global", 1, 1, []Arg{NewArg("args")}, "", func(interp *Interp, frame *Frame) string {
		return ""
	}})

	interp.AddCommand(ns_name, Command{"if", 1, 1, []Arg{NewArg("args")}, "", func(interp *Interp, frame *Frame) string {
		return ""
	}})

	interp.AddCommand(ns_name, Command{"proc", 3, 3, []Arg{NewArg("name"), NewArg("args"), NewArg("body")}, "", func(interp *Interp, frame *Frame) string {
		name, _ := frame.GetValue("name")
		args, _ := frame.GetValue("args")
		body, _ := frame.GetValue("body")

		_, words, _ := ParseWords(args.String())

		num_args := words.Len()
		min_args := num_args

		s_args := make([]Arg, num_args)

		i := 0
		for e := words.Front(); e != nil; e = e.Next() {
			_, arg_words, _ := ParseWords(e.Value.(Word).String())
			arg_name := arg_words.Remove(arg_words.Front()).(Word).String()

			if arg_words.Len() == 0 {
				s_args[i] = NewArg(arg_name)
			} else {
				arg_value := Value(arg_words.Remove(arg_words.Front()).(Word).String())
				s_args[i] = NewArgDefault(arg_name, &arg_value)
				if min_args == num_args {
					min_args = i
				}
			}

			i++
		}

		interp.AddCommand(ns_name, Command{name.String(), num_args, min_args, s_args, body.String(), nil})
		return ""
	}})

	interp.AddCommand(ns_name, Command{"puts", 1, 1, []Arg{NewArg("string")}, "", func(interp *Interp, frame *Frame) string {
		str, _ := frame.GetValue("string")
		fmt.Println(str.String())
		return ""
	}})

	interp.AddCommand(ns_name, Command{"pwd", 0, 0, []Arg{}, "", func(interp *Interp, frame *Frame) string {
		cwd, _ := os.Getwd()
		return cwd
	}})

	interp.AddCommand(ns_name, Command{"set", 2, 1, []Arg{NewArg("varName"), NewArgDefault("newValue", nil)}, "", func(interp *Interp, frame *Frame) string {
		var ok bool
		var value *Value

		varName, _ := frame.GetValue("varName")
		value, _ = frame.GetValue("newValue")

		level := frame.GetParentLevel()
		frame, error := interp.stack.GetFrame(level)

		if error != nil {
			fmt.Println(error)
			os.Exit(1)
		}

		if value != nil {
			frame.SetValue(varName.String(), value)
		} else {
			value, ok = frame.GetValue(varName.String())

			if !ok {
				fmt.Printf("can't read \"%s\": no such variable\n", varName)
				os.Exit(1)
			}
		}

		return value.String()
	}})

	interp.AddCommand(ns_name, Command{"uplevel", 2, 2, []Arg{NewArg("level"), NewArg("arg")}, "", func(interp *Interp, frame *Frame) string {
		mylevel := frame.GetLevel()

		level, _ := frame.GetValue("level")
		arg, _ := frame.GetValue("arg")

		if level.Len() >= 2 && level.String()[0] == '#' {
			level = StringToValueP(level.String()[1:])
		} else {
			level = IntToValueP(mylevel - level.Int())
		}

		other_frame, err := interp.stack.GetFrame(level.Int())

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		return other_frame.Eval(interp, arg.String())
	}})

	interp.AddCommand(ns_name, Command{"upvar", 3, 3, []Arg{NewArg("level"), NewArg("otherVar"), NewArg("localVar")}, "", func(interp *Interp, frame *Frame) string {
		mylevel := frame.GetLevel()

		level, _ := frame.GetValue("level")
		otherVar, _ := frame.GetValue("otherVar")
		localVar, _ := frame.GetValue("localVar")

		if level.Len() >= 2 && level.String()[0] == '#' {
			level = StringToValueP(level.String()[1:])
		} else {
			level = IntToValueP(mylevel - level.Int())
		}

		other_frame, err := interp.stack.GetFrame(level.Int())

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		otherValue, _ := other_frame.GetValue(otherVar.String())
		frame.SetValue(localVar.String(), otherValue)

		return ""
	}})
}

func (interp *Interp) FindCommand(name string, words *list.List) (cmd Command, error error) {
	ns := interp.namespaces["::"]

	cmd, error = ns.FindCommand(name, words)
	if error != nil {
		return
	}

	ok, error := cmd.ValidateArgs(words)
	if !ok {
		return
	}

	return
}

func ParseDoubleQuoteWord(script string) (ok bool, word Word, remainder string) {
	script = InterpSkipWhiteSpace(script)

	ok = false
	remainder = script

	loc := regexp.MustCompile("^\"([^\"])*\"").FindStringIndex(script)
	if loc == nil {
		return
	}

	ok = true
	remainder = InterpSkipWhiteSpace(string(script[loc[1]:]))
	word = DoubleQuoteWord(script[loc[0]+1 : loc[1]-1])

	return
}

func ParseBraceWord(script string) (ok bool, word Word, remainder string) {
	script = InterpSkipWhiteSpace(script)

	ok = false
	remainder = script

	loc := regexp.MustCompile("^{(({[^}]*})|[^}])*}").FindStringIndex(script)
	if loc == nil {
		return
	}

	ok = true
	word = BraceWord(script[loc[0]+1 : loc[1]-1])
	remainder = InterpSkipWhiteSpace(string(script[loc[1]:]))

	return
}

func ParseWord(script string) (ok bool, word Word, remainder string) {
	script = InterpSkipWhiteSpace(script)

	ok, word, remainder = ParseDoubleQuoteWord(script)
	if ok {
		return
	}

	ok, word, remainder = ParseBraceWord(script)
	if ok {
		return
	}

	ok = false
	remainder = script

	loc := regexp.MustCompile("^[^ \t\v\f\r\n;]+").FindStringIndex(script)
	if loc == nil {
		return
	}

	ok = true
	word = LiteralWord(script[loc[0]:loc[1]])
	remainder = InterpSkipWhiteSpace(string(script[loc[1]:]))

	return
}

func ParseWords(script string) (ok bool, words *list.List, remainder string) {
	var word Word

	remainder = script
	words = list.New()

	for len(remainder) != 0 {
		if remainder[0] == '\n' || remainder[0] == ';' {
			break
		}

		ok, word, remainder = ParseWord(remainder)
		if !ok {
			fmt.Printf("error parsing next word from script \"%s\"\n", remainder)
			os.Exit(1)
		}

		words.PushBack(word)
	}

	return true, words, remainder
}

func ParseComment(script string) (ok bool, comment string, remainder string) {
	script = InterpSkipWhiteSpace(script)

	ok = false
	comment = ""
	remainder = script

	loc := regexp.MustCompile("^#[^\n]*\n").FindStringIndex(script)
	if loc == nil {
		return
	}

	ok = true
	comment = script[loc[0]:loc[1]]
	remainder = script[loc[1]:]

	return
}

func InterpSkipWhiteSpace(script string) (remainder string) {
	remainder = script

	loc := regexp.MustCompile("^[ \t\v\r\f]+").FindStringIndex(script)
	if loc == nil {
		return
	}

	remainder = script[loc[1]:]
	return
}

func (interp *Interp) Eval(script string) string {
	frame := interp.stack.PushFrame()
	retval := frame.Eval(interp, script)
	interp.stack.PopFrame()
	return retval
}

func (interp *Interp) EvalWords(words *list.List) string {
	var retval string

	frame := interp.stack.PeekFrame()
	words = frame.SubstituteWords(interp, words)

	new_frame := interp.stack.PushFrame()

	name := words.Remove(words.Front()).(Word).String()
	cmd, error := interp.FindCommand(name, words)

	if error != nil {
		fmt.Println(error)
		os.Exit(1)
	}

	new_frame.BindArguments(cmd, words)

	if cmd.native_body == nil {
		retval = new_frame.Eval(interp, cmd.body)
	} else {
		retval = cmd.native_body(interp, new_frame)
	}

	interp.stack.PopFrame()

	return retval
}

func usage() {
	fmt.Printf("%s: FILE\n", os.Args[0])
}

func main() {
	if len(os.Args) != 2 {
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
