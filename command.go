// Niels Widger
// Time-stamp: <01 Aug 2013 at 19:43:37 by nwidger on macros.local>

package gotcl

import (
	"container/list"
	"errors"
	"fmt"
)

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

	if words.Len() < cmd.min_args || (cmd.num_args != -1 && words.Len() > cmd.num_args) {
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
