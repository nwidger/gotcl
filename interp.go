// Niels Widger
// Time-stamp: <02 Aug 2013 at 21:03:18 by nwidger on macros.local>

package gotcl

import (
	"container/list"
	"fmt"
	"os"
)

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

func (interp *Interp) AddBuiltinCommand(ns_name string, name string, args string, body func(*Interp, *Frame) string) bool {
	return interp.AddCommand_Aux(ns_name, name, args, "", body)
}

func (interp *Interp) AddCommand(ns_name string, name string, args string, body string) bool {
	return interp.AddCommand_Aux(ns_name, name, args, body, nil)
}

func (interp *Interp) AddCommand_Aux(ns_name string, name string, args string, body string, native_body func(*Interp, *Frame) string) bool {
	ns, ok := interp.namespaces[ns_name]

	if !ok {
		fmt.Printf("can't create procedure \"%v::%v\": unknown namespace\n", ns_name, name)
		os.Exit(1)
	}

	_, words, _ := ParseWords(args)

	num_args := words.Len()
	min_args := num_args

	s_args := make([]Arg, num_args)

	for i, e := 0, words.Front(); e != nil; i, e = i+1, e.Next() {
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

		if e.Next() == nil && arg_name == "args" {
			num_args = -1
		}
	}

	cmd := Command{name, num_args, min_args, s_args, body, native_body}

	return ns.AddCommand(cmd)
}

func (interp *Interp) AddBuiltinCommands() {
	ns_name := "::"

	// cd
	interp.AddBuiltinCommand(ns_name, "cd", "dir", func(interp *Interp, frame *Frame) string {
		dir, _ := frame.GetValue("dir")
		os.Chdir(dir.String())
		return ""
	})

	// eval
	interp.AddBuiltinCommand(ns_name, "eval", "script", func(interp *Interp, frame *Frame) string {
		script, _ := frame.GetValue("script")
		return interp.Eval(script.String())
	})

	// global
	interp.AddBuiltinCommand(ns_name, "global", "args", func(interp *Interp, frame *Frame) string {
		return ""
	})

	// if
	interp.AddBuiltinCommand(ns_name, "if", "args", func(interp *Interp, frame *Frame) string {
		return ""
	})

	// list
	interp.AddBuiltinCommand(ns_name, "list", "args", func(interp *Interp, frame *Frame) string {
		return ""
	})

	// proc
	interp.AddBuiltinCommand(ns_name, "proc", "name args body", func(interp *Interp, frame *Frame) string {
		name, _ := frame.GetValue("name")
		args, _ := frame.GetValue("args")
		body, _ := frame.GetValue("body")
		interp.AddCommand(ns_name, name.String(), args.String(), body.String())
		return ""
	})

	// puts
	interp.AddBuiltinCommand(ns_name, "puts", "string", func(interp *Interp, frame *Frame) string {
		str, _ := frame.GetValue("string")
		fmt.Println(str.String())
		return ""
	})

	// pwd
	interp.AddBuiltinCommand(ns_name, "pwd", "", func(interp *Interp, frame *Frame) string {
		cwd, _ := os.Getwd()
		return cwd
	})

	// set
	interp.AddBuiltinCommand(ns_name, "set", "varName newValue", func(interp *Interp, frame *Frame) string {
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
	})

	// uplevel
	interp.AddBuiltinCommand(ns_name, "uplevel", "level arg", func(interp *Interp, frame *Frame) string {
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
	})

	// upvar
	interp.AddBuiltinCommand(ns_name, "upvar", "level otherVar localVar", func(interp *Interp, frame *Frame) string {
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
	})
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
