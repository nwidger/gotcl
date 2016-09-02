package gotcl

import (
	"strconv"
	"unicode"
)

func backslashNewlineSubstOnce(r []rune) ([]rune, bool) {
	if len(r) < 2 || r[0] != '\\' {
		return r, false
	}

	// \<newline>whiteSpace
	//
	// A single space character replaces the backslash, newline,
	// and all spaces and tabs after the newline. This backslash
	// sequence is unique in that it is replaced in a separate
	// pre-pass before the command is actually parsed. This means
	// that it will be replaced even when it occurs between
	// braces, and the resulting space will be treated as a word
	// separator if it is not in braces or quotes.
	if r[1] != '\n' {
		return r, false
	}

	var i int

	for i = 2; i < len(r); i++ {
		if !unicode.IsSpace(r[i]) {
			break
		}
	}

	return append([]rune{' '}, r[i:]...), true
}

func BackslashNewlineSubst(r []rune) ([]rune, bool) {
	origLength := len(r)

	for i := 0; i < len(r); i++ {
		if r[i] != '\\' {
			continue
		}
		nb, ok := backslashNewlineSubstOnce(r[i:])
		if ok {
			r = append(r[:i], nb...)
		}
	}

	return r, len(r) != origLength
}

// If a backslash (“\”) appears within a word then backslash
// substitution occurs.  In all cases but those described below the
// backslash is dropped and the following character is treated as an
// ordinary character and included in the word.  This allows
// characters such as double quotes, close brackets, and dollar signs
// to be included in words without triggering special processing. The
// following table lists the backslash sequences that are handled
// specially, along with the value that replaces each sequence.
func backslashSubstOnce(r []rune) ([]rune, bool) {
	if len(r) < 2 || r[0] != '\\' {
		return r, false
	}

	first := r[1]

	// \a
	//     Audible alert (bell) (Unicode U+000007).
	// \b
	//     Backspace (Unicode U+000008).
	// \f
	//     Form feed (Unicode U+00000C).
	// \n
	//     Newline (Unicode U+00000A).
	// \r
	//     Carriage-return (Unicode U+00000D).
	// \t
	//     Tab (Unicode U+000009).
	// \v
	//     Vertical tab (Unicode U+00000B).
	// \\
	//     Backslash (“\”).
	switch first {
	case 'a':
		return append([]rune{'\a'}, r[2:]...), true
	case 'b':
		return append([]rune{'\b'}, r[2:]...), true
	case 'f':
		return append([]rune{'\f'}, r[2:]...), true
	case 'n':
		return append([]rune{'\n'}, r[2:]...), true
	case 'r':
		return append([]rune{'\r'}, r[2:]...), true
	case 't':
		return append([]rune{'\t'}, r[2:]...), true
	case 'v':
		return append([]rune{'\v'}, r[2:]...), true
	case '\\':
		return append([]rune{'\\'}, r[2:]...), true
	}

	// \ooo
	//
	//     The digits ooo (one, two, or three of them) give a
	//     eight-bit octal value for the Unicode character that
	//     will be inserted, in the range 000–377 (i.e., the range
	//     U+000000–U+0000FF). The parser will stop just before
	//     this range overflows, or when the maximum of three
	//     digits is reached. The upper bits of the Unicode
	//     character will be 0.
	//
	//     The range U+010000–U+10FFFD is reserved for the future.
	//
	// \xhh
	//
	//     The hexadecimal digits hh (one or two of them) give an
	//     eight-bit hexadecimal value for the Unicode character
	//     that will be inserted. The upper bits of the Unicode
	//     character will be 0 (i.e., the character will be in the
	//     range U+000000–U+0000FF).
	//
	// \uhhhh
	//
	//     The hexadecimal digits hhhh (one, two, three, or four
	//     of them) give a sixteen-bit hexadecimal value for the
	//     Unicode character that will be inserted. The upper bits
	//     of the Unicode character will be 0 (i.e., the character
	//     will be in the range U+000000–U+00FFFF).
	//
	// \Uhhhhhhhh
	//
	//     The hexadecimal digits hhhhhhhh (one up to eight of
	//     them) give a twenty-one-bit hexadecimal value for the
	//     Unicode character that will be inserted, in the range
	//     U+000000–U+10FFFF. The parser will stop just before
	//     this range overflows, or when the maximum of eight
	//     digits is reached. The upper bits of the Unicode
	//     character will be 0.
	//
	//     The range U+010000–U+10FFFD is reserved for the future.
	var (
		appendRune  func(rune) bool
		validChar   func(int64) bool
		maxLen, idx int
		base        int
	)

	switch first {
	case '0', '1', '2', '3',
		'4', '5', '6', '7':
		idx = 1
		base = 8
		appendRune = func(r rune) bool {
			return '0' <= r && r <= '7'
		}

		maxLen = 3
		validChar = func(val int64) bool {
			return val >= 0x000000 && val <= 0x0000ff
		}
	case 'x', 'u', 'U':
		idx = 2
		base = 16
		appendRune = func(r rune) bool {
			return ('0' <= r && r <= '9') ||
				('a' <= r && r <= 'f') ||
				('A' <= r && r <= 'F')
		}

		switch first {
		case 'x':
			maxLen = 2
			validChar = func(val int64) bool {
				return val >= 0x000000 && val <= 0x0000ff
			}
		case 'u':
			maxLen = 4
			validChar = func(val int64) bool {
				return val >= 0x000000 && val <= 0x00ffff
			}
		case 'U':
			maxLen = 8
			validChar = func(val int64) bool {
				return (val >= 0x000000 && val <= 0x00ffff) ||
					(val >= 0x10fffe && val <= 0x10ffff)
			}
		}
	default:
		maxLen = -1
	}

	var buf []rune
	for i := idx; i < len(r) && len(buf) < maxLen; i++ {
		if !appendRune(r[i]) {
			break
		}
		buf = append(buf, r[i])
	}
	for ; len(buf) > 0; buf = buf[:len(buf)-1] {
		val, err := strconv.ParseInt(string(buf), base, 64)
		if err != nil || !validChar(val) {
			continue
		}
		return append([]rune{rune(val)}, r[len(buf)+idx:]...), true
	}

	// Backslash substitution is not performed on words enclosed
	// in braces, except for backslash-newline as described above.

	// In all cases but those described below the backslash is
	// dropped and the following character is treated as an
	// ordinary character and included in the word.
	return append([]rune{r[1]}, r[2:]...), true
}

func BackslashSubst(r []rune) ([]rune, bool) {
	origLength := len(r)

	for i := 0; i < len(r); i++ {
		if r[i] != '\\' {
			continue
		}
		nb, ok := backslashSubstOnce(r[i:])
		if ok {
			r = append(r[:i], nb...)
		}
	}

	return r, len(r) != origLength
}
