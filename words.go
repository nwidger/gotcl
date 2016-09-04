package gotcl

import (
	"fmt"
	"unicode"
)

func ParseBareWord(r []rune) ([]rune, int, error) {
	var (
		idx  int
		prev rune
	)

	for idx = 1; idx < len(r); idx++ {
		c := r[idx]

		if unicode.IsSpace(c) {
			break
		}

		if c == '[' && prev != '\\' {
			_, size, err := ParseBracketWord(r[idx:])
			if err != nil {
				return nil, 0, err
			}
			idx += size
		}

		prev = c
	}
	if idx > len(r) {
		idx = len(r)
	}

	return r[:idx], idx + 1, nil
}

// If the first character of a word is double-quote (“"”) then the
// word is terminated by the next double-quote character. If
// semi-colons, close brackets, or white space characters (including
// newlines) appear between the quotes then they are treated as
// ordinary characters and included in the word. Command substitution,
// variable substitution, and backslash substitution are performed on
// the characters between the quotes as described below. The
// double-quotes are not retained as part of the word.
func ParseDoubleQuoteWord(r []rune) ([]rune, int, error) {
	if len(r) == 0 || r[0] != '"' {
		return nil, 0, fmt.Errorf("word does not start with double-quote")
	}
	if len(r) < 2 {
		return nil, 0, fmt.Errorf("unterminated double-quote")
	}

	var prev rune
	closed := false
	idx := 1

	for i := idx; i < len(r); i++ {
		c := r[i]
		if c == '"' && prev != '\\' {
			closed = true
			idx = i
			break
		}
		prev = c
	}
	if idx > len(r) {
		idx = len(r)
	}

	if !closed {
		return nil, 0, fmt.Errorf("unterminated double-quote word")
	}

	return r[1:idx], idx + 1, nil
}

// If the first character of a word is an open brace (“{”) and rule
// [5] does not apply, then the word is terminated by the matching
// close brace (“}”). Braces nest within the word: for each additional
// open brace there must be an additional close brace (however, if an
// open brace or close brace within the word is quoted with a
// backslash then it is not counted in locating the matching close
// brace). No substitutions are performed on the characters between
// the braces except for backslash-newline substitutions described
// below, nor do semi-colons, newlines, close brackets, or white space
// receive any special interpretation. The word will consist of
// exactly the characters between the outer braces, not including the
// braces themselves.
func ParseBraceWord(r []rune) ([]rune, int, error) {
	if len(r) == 0 || r[0] != '{' {
		return nil, 0, fmt.Errorf("word does not start with open brace")
	}
	if len(r) < 2 {
		return nil, 0, fmt.Errorf("unterminated brace word")
	}

	if len(r) >= 4 &&
		r[0] == '{' &&
		r[1] == '*' &&
		r[2] == '}' &&
		!unicode.IsSpace(r[3]) {
		return ParseArgumentExpansionWord(r)
	}

	var prev rune
	open := 1
	idx := 1

	for i := idx; i < len(r); i++ {
		c := r[i]
		switch {
		case c == '{' && prev != '\\':
			open++
		case c == '}' && prev != '\\':
			open--
		}
		if open == 0 {
			idx = i
			break
		}
		prev = c
	}

	if open != 0 {
		return nil, 0, fmt.Errorf("unterminated brace word")
	}

	return r[1:idx], idx + 1, nil
}

// If a word starts with the string “{*}” followed by a non-whitespace
// character, then the leading “{*}” is removed and the rest of the
// word is parsed and substituted as any other word. After
// substitution, the word is parsed as a list (without command or
// variable substitutions; backslash substitutions are performed as is
// normal for a list and individual internal words may be surrounded
// by either braces or double-quote characters), and its words are
// added to the command being substituted. For instance, “cmd a {*}{b
// [c]} d {*}{$e f {g h}}” is equivalent to “cmd a b {[c]} d {$e} f {g
// h}”.
func ParseArgumentExpansionWord(r []rune) ([]rune, int, error) {
	if len(r) < 4 ||
		r[0] != '{' ||
		r[1] != '*' ||
		r[2] != '}' ||
		unicode.IsSpace(r[3]) {
		return nil, 0, fmt.Errorf("word does not start with '{*}' followed by a non-whitespace character")
	}

	nb, size, err := ParseWord(r[3:])
	if err != nil {
		return nil, 0, err
	}

	return nb, size + 3, nil
}

// If a word contains an open bracket (“[”) then Tcl performs command
// substitution. To do this it invokes the Tcl interpreter recursively
// to process the characters following the open bracket as a Tcl
// script. The script may contain any number of commands and must be
// terminated by a close bracket (“]”). The result of the script
// (i.e. the result of its last command) is substituted into the word
// in place of the brackets and all of the characters between
// them. There may be any number of command substitutions in a single
// word. Command substitution is not performed on words enclosed in
// braces.
func ParseBracketWord(r []rune) ([]rune, int, error) {
	if len(r) == 0 || r[0] != '[' {
		return nil, 0, fmt.Errorf("word does not start with open bracket")
	}
	if len(r) < 2 {
		return nil, 0, fmt.Errorf("unterminated bracket word")
	}

	var prev rune
	open := 1
	idx := 1

	for i := idx; i < len(r); i++ {
		c := r[i]
		switch {
		case c == '[' && prev != '\\':
			open++
		case c == ']' && prev != '\\':
			open--
		}
		if open == 0 {
			idx = i
			break
		}
		prev = c
	}

	if open != 0 {
		return nil, 0, fmt.Errorf("unterminated bracket word")
	}

	return r[1:idx], idx + 1, nil
}

// A Tcl script is a string containing one or more
// commands. Semi-colons and newlines are command separators unless
// quoted as described below. Close brackets are command terminators
// during command substitution (see below) unless quoted.
//
// Words of a command are separated by white space (except for
// newlines, which are command separators).
func ParseWords(r []rune) ([][]rune, error) {
	nb, ok := BackslashNewlineSubst(r)
	if ok {
		r = nb
	}

	ws := [][]rune{}

	var prev rune
	idx := 0

	for i := idx; i < len(r); i++ {
		c := r[i]

		if (c == ';' || c == '\n') && prev != '\\' {
			break
		}

		if unicode.IsSpace(c) {
			continue
		}

		w, size, err := ParseWord(r[i:])
		if err != nil {
			return nil, err
		}
		ws = append(ws, w)
		i += size

		prev = c
	}

	return ws, nil
}

func ParseWord(r []rune) ([]rune, int, error) {
	switch {
	case r[0] == '"':
		return ParseDoubleQuoteWord(r)
	case r[0] == '{':
		return ParseBraceWord(r)
	default:
		return ParseBareWord(r)
	}
}
