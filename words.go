package gotcl

import (
	"fmt"
	"unicode"
)

func ParseWord(r []rune) ([]rune, int, error) {
	idx := 1

	for i := idx; i < len(r); i++ {
		if unicode.IsSpace(r[i]) {
			idx = i
			break
		}
	}

	return r[:idx], idx + 1, nil
}

func ParseDoubleQuoteWord(r []rune) ([]rune, int, error) {
	if len(r) < 2 || r[0] != '"' {
		return nil, 0, fmt.Errorf("word does not start with double-quote")
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

	if !closed {
		return nil, 0, fmt.Errorf("unterminated double-quote word")
	}

	return r[1:idx], idx + 1, nil
}

func ParseBraceWord(r []rune) ([]rune, int, error) {
	if len(r) < 2 || r[0] != '{' {
		return nil, 0, fmt.Errorf("word does not start with open brace")
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

func ParseArgumentExpansionWord(r []rune) ([]rune, int, error) {
	return r, 0, nil
}

func ParseWords(r []rune) ([][]rune, error) {
	ws := [][]rune{}
	idx := 0

	for i := idx; i < len(r); i++ {
		c := r[i]
		switch {
		case c == '"' && i+1 < len(r):
			w, size, err := ParseDoubleQuoteWord(r[i:])
			if err != nil {
				return nil, err
			}
			ws = append(ws, w)
			i += size
		case c == '{' && i+1 < len(r):
			w, size, err := ParseBraceWord(r[i:])
			if err != nil {
				return nil, err
			}
			ws = append(ws, w)
			i += size
		case unicode.IsSpace(c):
			continue
		default:
			w, size, err := ParseWord(r[i:])
			if err != nil {
				return nil, err
			}
			ws = append(ws, w)
			i += size
		}
	}

	return ws, nil
}
