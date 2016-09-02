package gotcl

import "unicode"

type litWord []byte

type cmdWord []byte

type varSubstWord []byte

// $name
//
//  Name is the name of a scalar variable; the name is a sequence of
//  one or more characters that are a letter, digit, underscore, or
//  namespace separators (two or more colons). Letters and digits are
//  only the standard ASCII ones (0–9, A–Z and a–z).
type scalarVarSubstWord []byte

// $name(index)
//
//  Name gives the name of an array variable and index gives the name
//  of an element within that array. Name must contain only letters,
//  digits, underscores, and namespace separators, and may be an empty
//  string. Letters and digits are only the standard ASCII ones (0–9,
//  A–Z and a–z). Command substitutions, variable substitutions, and
//  backslash substitutions are performed on the characters of index.
type arrayElemSubstWord struct {
	name  []byte
	index []byte
}

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
type scalarVarOrArrayElemSubstWord []byte

type Word []byte

type Command []Word

const CommentRune = '#'

func IsCommandSep(r rune) bool {
	return r == '\n' || r == ';'
}

func IsWordSep(r rune) bool {
	return r != '\n' && unicode.IsSpace(r)
}
