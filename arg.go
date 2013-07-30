// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:25:52 by nwidger on macros.local>

package gotcl

import (
	"fmt"
)

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
