// Niels Widger
// Time-stamp: <02 Aug 2013 at 21:21:36 by nwidger on macros.local>

package gotcl

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
	"strings"
)

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
	len := len(cmd.args)

	for i, arg := range cmd.args {
		name := arg.GetName()
		value := arg.GetValue()

		if i == (len-1) && name == "args" {
			tmp := []string{}

			for words.Len() != 0 {
				word := words.Remove(words.Front()).(Word).String()
				tmp = append(tmp, word)
			}

			value = StringToValueP(strings.Join(tmp, " "))
		} else {
			if words.Len() != 0 {
				word := Value(words.Remove(words.Front()).(Word).String())
				value = &word
			}
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
	switch word.(type) {
	case BraceWord:
		return word
	}

	newword := LiteralWord("")

	for len(word.String()) != 0 {
		cmd_ok, cmd_part, cmd_word := frame.SubstituteCommand(interp, word)
		var_ok, var_part, var_word := frame.SubstituteVariable(interp, word)

		if cmd_ok && (!var_ok || cmd_word.Len() >= var_word.Len()) {
			newword = LiteralWord(newword.String() + cmd_part.String())
			word = cmd_word
		} else if var_ok && (!cmd_ok || var_word.Len() >= cmd_word.Len()) {
			newword = LiteralWord(newword.String() + var_part.String())
			word = var_word
		} else {
			newword = LiteralWord(newword.String() + word.String())
			word = LiteralWord("")
		}
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
