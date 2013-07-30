// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:31:41 by nwidger on macros.local>

package gotcl

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
)

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
