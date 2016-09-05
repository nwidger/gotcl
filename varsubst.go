package gotcl

import "fmt"

type variableFunc func(name, index []rune) (value []rune, err error)

// If a word contains a dollar-sign (“$”) followed by one of the forms
// described below, then Tcl performs variable substitution: the
// dollar-sign and the following characters are replaced in the word
// by the value of a variable. Variable substitution may take any of
// the following forms:
//
// $name
//
// Name is the name of a scalar variable; the name is a sequence of
// one or more characters that are a letter, digit, underscore, or
// namespace separators (two or more colons). Letters and digits are
// only the standard ASCII ones (0–9, A–Z and a–z).
//
// $name(index)
//
// Name gives the name of an array variable and index gives the name
// of an element within that array. Name must contain only letters,
// digits, underscores, and namespace separators, and may be an empty
// string. Letters and digits are only the standard ASCII ones (0–9,
// A–Z and a–z). Command substitutions, variable substitutions, and
// backslash substitutions are performed on the characters of index.
//
// ${name}
//
// Name is the name of a scalar variable or array element. It may
// contain any characters whatsoever except for close braces. It
// indicates an array element if name is in the form
// “arrayName(index)” where arrayName does not contain any open
// parenthesis characters, “(”, or close brace characters, “}”, and
// index can be any sequence of characters except for close brace
// characters. No further substitutions are performed during the
// parsing of name.
//
// There may be any number of variable substitutions in a single
// word. Variable substitution is not performed on words enclosed in
// braces.
//
// Note that variables may contain character sequences other than
// those listed above, but in that case other mechanisms must be used
// to access them (e.g., via the set command's single-argument form).
func ParseVarName(r []rune) ([]rune, []rune, int, error) {
	var (
		name  []rune
		index []rune
		idx   int
		prev  rune
	)

	if len(r) == 0 || r[0] != '$' {
		return nil, nil, 0, fmt.Errorf("must start with '$'")
	}

	start, closed, brace, array := 1, true, false, false
	if r[1] == '{' {
		start, closed, brace = 2, false, true
	}

	colons := 0
	nameChar := func(r, prev rune) bool {
		if array {
			closed = r == ')' && prev != '\\'
			return !closed
		}
		if brace {
			closed = r == '}'
			return !closed
		}

		if r == ':' {
			colons++
			return true
		} else {
			if prev == ':' && colons < 2 {
				return false
			}
			colons = 0
		}
		return r == '_' ||
			('0' <= r && r <= '9') ||
			('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z')
	}

	for idx = start; idx < len(r); idx++ {
		if !brace && !array && r[idx] == '(' {
			name = r[start:idx]
			start, closed, array = idx+1, false, true
		}

		if !nameChar(r[idx], prev) {
			break
		}

		prev = r[idx]
	}
	if !closed {
		if brace {
			return nil, nil, 0, fmt.Errorf("missing close-brace for variable name")
		}
		if array {
			return nil, nil, 0, fmt.Errorf("missing )")
		}
	}

	if !array {
		name = r[start:idx]
	} else {
		index = r[start:idx]
	}

	if brace || array {
		idx++
	}

	return name, index, idx, nil
}

func ParseVar(r []rune, variable variableFunc) ([]rune, int, error) {
	name, index, size, err := ParseVarName(r)
	if err != nil {
		return nil, 0, err
	}

	value, err := variable(name, index)
	if err != nil {
		return nil, 0, err
	}

	return append(value, r[size:]...), len(value), nil
}

func VarSubst(r []rune, variable variableFunc) ([]rune, error) {
	for i := 0; i < len(r); i++ {
		if r[i] != '$' {
			continue
		}
		nb, length, err := ParseVar(r[i:], variable)
		if err != nil {
			return nil, err
		}
		r = append(r[:i], nb...)
		i += length
	}

	return r, nil
}
