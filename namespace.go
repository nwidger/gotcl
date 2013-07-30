// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:25:29 by nwidger on macros.local>

package gotcl

import (
	"container/list"
	"errors"
	"fmt"
)

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
