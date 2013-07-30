// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:31:00 by nwidger on macros.local>

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
