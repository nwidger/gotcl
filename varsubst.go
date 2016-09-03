package gotcl

type substFuncFlags int

const (
	SUBST_COMMANDS substFuncFlags = iota
	SUBST_VARIABLES
	SUBST_BACKSLASHES

	SUBST_ALL = SUBST_COMMANDS | SUBST_VARIABLES | SUBST_BACKSLASHES
)

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
func VarSubstOnce(r []rune, variable variableFunc) ([]rune, int, error) {
	if len(r) < 2 || r[0] != '$' {
		return nil, 0, nil
	}

	brace := false
	start := 1
	nameChar := func(r rune) bool {
		return r == '_' || r == ':' ||
			('0' <= r && r <= '9') ||
			('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z')
	}

	if r[1] == '{' {
		brace = true
		start = 2
		nameChar = func(r rune) bool {
			return r != '}'
		}
	}

	var idx int
	for idx = start; idx < len(r); idx++ {
		if r[idx] == '(' {
			break
		}
		if !nameChar(r[idx]) {
			break
		}
	}

	var (
		name  []rune
		index []rune
	)

	name = r[start:idx]

	if len(r) > idx && r[idx] == '(' {
		var prev rune
		start = idx + 1
		for idx = start; idx < len(r); idx++ {
			if r[idx] == ')' && (brace || prev != '\\') {
				break
			}
			prev = r[idx]
		}
		index = r[start:idx]
		idx++

		if !brace {
			// TODO: perform command, variable and
			// backslash substitution on index
		}
	}

	if brace {
		idx++
	}

	value, err := variable(name, index)
	if err != nil {
		return nil, 0, err
	}

	return append(value, r[idx:]...), len(value), nil
}

func VarSubst(r []rune, variable variableFunc) ([]rune, error) {
	for i := 0; i < len(r); i++ {
		if r[i] != '$' {
			continue
		}
		nb, length, err := VarSubstOnce(r[i:], variable)
		if err != nil {
			return nil, err
		}
		r = append(r[:i], nb...)
		i += length
	}

	return r, nil
}
