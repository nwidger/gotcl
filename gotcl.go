package gotcl

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type TermType int

const (
	TermQuote        TermType = 1 << iota // double-quote '"'
	TermSpace                             // whitespace
	TermCommandEnd                        // newline or semicolon
	TermCloseParen                        // close paren ')'
	TermCloseBracket                      // close bracket ']'
)

type SubstType int

const (
	SubstCommands SubstType = 1 << iota
	SubstVariables
	SubstBackslashes

	SubstAll SubstType = SubstCommands | SubstVariables | SubstBackslashes
)

func ParseAllWhiteSpace(r []rune) ([]rune, int, error) {
	return parseWhiteSpace(r, true)
}

func ParseWhiteSpace(r []rune) ([]rune, int, error) {
	return parseWhiteSpace(r, false)
}

func parseWhiteSpace(r []rune, newlines bool) ([]rune, int, error) {
	var (
		idx int
	)
	for idx = 0; idx < len(r); idx++ {
		var next rune
		c := r[idx]
		if idx < len(r)-1 {
			next = r[idx+1]
		}
		if c == '\t' || c == '\v' || c == '\f' || c == '\r' || c == ' ' ||
			(c == '\n' && newlines) ||
			(c == '\\' && next == '\n') {
			continue
		}
		if c == '\\' && next == '\n' {
			_, size, err := ParseBackslashNewlineToken(r[idx:])
			if err != nil {
				return nil, 0, err
			}
			if size > 0 {
				idx += size - 1
			}
			continue
		}
		break
	}
	return r[:idx], idx, nil
}

func ParseComment(r []rune) ([]rune, int, error) {
	var (
		idx int
	)
	for idx = 0; idx < len(r); idx++ {
		_, size, err := ParseAllWhiteSpace(r[idx:])
		if err != nil {
			return nil, 0, err
		}
		if size > 0 {
			idx += size
		}
		if r[idx] != '#' {
			break
		}
		for ; idx < len(r); idx++ {
			var next rune
			c := r[idx]
			if idx < len(r)-1 {
				next = r[idx+1]
			}
			if c != '\\' {
				if next == '\n' {
					idx++
					break
				}
			} else {
				_, size, err := ParseWhiteSpace(r[idx:])
				if err != nil {
					return nil, 0, err
				}
				if size > 0 {
					idx += size - 1
				} else {
					_, size, err := ParseBackslashToken(r[idx:])
					if err != nil {
						return nil, 0, err
					}
					if size > 0 {
						idx += size - 1
					}
				}
			}
		}
	}
	return r[:idx], idx, nil
}

// words must only contain wordToken, simpleWordToken or
// expandWordToken
func ParseCommand(r []rune, nested bool) (words, int, error) {
	ws := words{}
	var (
		idx  int
		prev rune
	)
	_, size, err := ParseComment(r)
	if err != nil {
		return nil, 0, err
	}
	for idx = size; idx < len(r); idx++ {
		var next rune
		c := r[idx]
		if idx < len(r)-1 {
			next = r[idx+1]
		}
		if c == '\\' && next == '\n' {
			_, size, err := ParseBackslashNewlineToken(r[idx:])
			if err != nil {
				return nil, 0, err
			}
			if size > 0 {
				idx += size - 1
			}
			prev = r[idx]
			continue
		}
		if prev != '\\' && (c == ';' || c == '\n') {
			if !nested && len(ws) == 0 {
				continue
			}
			idx++
			break
		}
		if nested && c == ']' {
			break
		}
		if unicode.IsSpace(c) {
			prev = c
			continue
		}
		w, size, err := ParseWord(r[idx:], nested)
		if err != nil {
			return nil, 0, err
		}
		if size > 0 {
			idx += size - 1
		}
		ws = append(ws, w)
		prev = r[idx]
	}
	return ws, idx, nil
}

func ParseWord(r []rune, nested bool) (word, int, error) {
	if len(r) == 0 {
		return simpleWordToken(r), 0, nil
	}

	argumentExpansion := false
	if len(r) >= 4 &&
		r[0] == '{' &&
		r[1] == '*' &&
		r[2] == '}' &&
		!unicode.IsSpace(r[3]) {
		argumentExpansion = true
		r = r[3:]
	}

	switch r[0] {
	case '"':
		w, size, err := ParseQuotedStringWord(r, nested)
		if err != nil {
			return nil, 0, err
		}
		if argumentExpansion {
			return expandWordToken{w}, size + 3, nil
		}
		return w, size, nil
	case '{':
		w, size, err := ParseBracesWord(r, nested)
		if err != nil {
			return nil, 0, err
		}
		if argumentExpansion {
			return expandWordToken{w}, size + 3, nil
		}
		return w, size, nil
	}

	terminators := TermSpace | TermCommandEnd
	if nested {
		terminators |= TermCloseBracket
	}
	ts, size, err := ParseTokens(r, terminators, SubstAll)
	if err != nil {
		return nil, 0, err
	}

	var w word
	w = wordToken(ts)
	if len(ts) == 1 {
		if text, ok := ts[0].(textToken); ok {
			w = simpleWordToken(text)
		}
	}

	if argumentExpansion {
		return expandWordToken{w}, size + 3, nil
	}
	return w, size, nil
}

func ParseTextToken(r []rune, terminators TermType) (token, int, error) {
	var (
		idx  int
		prev rune
	)
outerLoop:
	for idx = 0; idx < len(r); idx++ {
		var next rune
		c := r[idx]
		if idx < len(r)-1 {
			next = r[idx+1]
		}
		switch {
		case (terminators&TermSpace) != 0 &&
			(prev != '\\' &&
				(c == '\t' || c == '\v' || c == '\f' || c == '\r' || c == ' ')):
			break outerLoop
		case (terminators&TermSpace) != 0 &&
			(c == '\\' && next == '\n'):
			break outerLoop
		case (terminators&TermCommandEnd) != 0 &&
			(prev != '\\' && c == ';' || prev != '\\' && c == '\n'):
			break outerLoop
		case (terminators&TermCloseBracket) != 0 &&
			prev != '\\' && c == ']':
			break outerLoop
		case prev != '\\' && c == '\n',
			prev != '\\' && c == '\\',
			prev != '\\' && c == '[',
			prev != '\\' && c == '$':
			break outerLoop
		}
		prev = c
	}
	return textToken(r[:idx]), idx, nil
}

// If the first character of a word is double-quote (“"”) then the
// word is terminated by the next double-quote character. If
// semi-colons, close brackets, or white space characters (including
// newlines) appear between the quotes then they are treated as
// ordinary characters and included in the word. Command substitution,
// variable substitution, and backslash substitution are performed on
// the characters between the quotes as described below. The
// double-quotes are not retained as part of the word.
func ParseQuotedStringWord(r []rune, nested bool) (word, int, error) {
	if len(r) == 0 || r[0] != '"' {
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

	if idx+1 < len(r) &&
		r[idx+1] != ';' &&
		r[idx+1] != '\n' &&
		(!nested || r[idx+1] != ']') &&
		!unicode.IsSpace(r[idx+1]) {
		return nil, 0, fmt.Errorf("extra characters after close-quote")
	}

	size := idx + 1
	ws, err := ParseQuotedStringTokens(simpleWordToken(r[:idx+1]), nested)
	if err != nil {
		return nil, 0, err
	}

	return ws, size, nil
}

func ParseQuotedStringTokens(r simpleWordToken, nested bool) (word, error) {
	if len(r) == 0 || r[0] != '"' {
		return nil, fmt.Errorf("word does not start with double-quote")
	}

	ts, _, err := ParseTokens(r[1:len(r)-1], TermQuote, SubstAll)
	if err != nil {
		return nil, err
	}

	return wordToken(ts), nil
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
func ParseBracesWord(r []rune, nested bool) (word, int, error) {
	if len(r) == 0 || r[0] != '{' {
		return nil, 0, fmt.Errorf("word does not start with open brace")
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

	if idx+1 < len(r) &&
		r[idx+1] != ';' &&
		r[idx+1] != '\n' &&
		(!nested || r[idx+1] != ']') &&
		!unicode.IsSpace(r[idx+1]) {
		return nil, 0, fmt.Errorf("extra characters after close-quote")
	}

	size := idx + 1
	ws, err := ParseBracesTokens(simpleWordToken(r[:idx+1]), nested)
	if err != nil {
		return nil, 0, err
	}

	return ws, size, nil
}

// must be '{' text '}'
func ParseBracesTokens(r simpleWordToken, nested bool) (word, error) {
	var (
		start int
		end   int
	)

	if len(r) == 0 || r[0] != '{' {
		return nil, fmt.Errorf("word does not start with open brace")
	}

	ws := wordToken{}

	for start, end = 1, 1; end < len(r)-1; end++ {
		c := r[end]
		if c == '\\' && end+1 < len(r)-1 && r[end+1] == '\n' {
			if end != start {
				ws = append(ws, textToken(r[start:end]))
			}
			tok, size, err := ParseBackslashNewlineToken(r[end:])
			if err != nil {
				return nil, err
			}
			ws = append(ws, tok)
			end += size
			start = end
		}
	}
	if end != start {
		ws = append(ws, textToken(r[start:end]))
	}

	return ws, nil
}

func ParseTokens(r []rune, terminators TermType, substs SubstType) (tokens, int, error) {
	ts := tokens{}

	var (
		idx  int
		prev rune
	)

	// construct wordToken composed of
	// [textToken|bsToken|commandToken|variableToken], terminated
	// by whitespace, semicolon, newline or an unescaped close
	// bracket '] if nested'
	for idx = 0; idx < len(r); idx++ {
		var next rune
		c := r[idx]
		if idx < len(r)-1 {
			next = r[idx+1]
		}
		if (terminators&TermCommandEnd) != 0 &&
			prev != '\\' && (c == ';' || c == '\n') {
			break
		}
		if (terminators&TermCloseBracket) != 0 &&
			prev != '\\' && c == ']' {
			break
		}
		if (terminators&TermSpace) != 0 &&
			(unicode.IsSpace(c) || c == '\\' && next == '\n') {
			break
		}
		var (
			tok  token
			size int
			err  error
		)
		switch {
		case c == '\\' && prev != '\\':
			if (substs & SubstBackslashes) == 0 {
				tok, size, err = textToken(r[idx:idx+1]), 1, nil
			} else {
				tok, size, err = ParseBackslashToken(r[idx:])
			}
		case c == '[' && prev != '\\':
			if (substs & SubstCommands) == 0 {
				tok, size, err = textToken(r[idx:idx+1]), 1, nil
			} else {
				_, size, err = ParseCommand(r[idx+1:], true)
				if err != nil {
					return nil, 0, err
				}
				size += 2
				tok = commandToken(r[idx : idx+size])
			}
		case c == '$' && prev != '\\':
			if (substs & SubstVariables) == 0 {
				tok, size, err = textToken(r[idx:idx+1]), 1, nil
			} else {
				tok, size, err = ParseVarNameToken(r[idx:])
			}
		default:
			tok, size, err = ParseTextToken(r[idx:], terminators)
		}
		if err != nil {
			return nil, 0, err
		}
		if size > 0 {
			idx += size - 1
		}
		ts = append(ts, tok)
		prev = r[idx]
	}

	return ts, idx, nil
}

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
func ParseVarNameToken(r []rune) (token, int, error) {
	var (
		name  []rune
		index []rune
		idx   int
		prev  rune
	)

	if len(r) == 0 || r[0] != '$' {
		return nil, 0, fmt.Errorf("must start with '$'")
	}

	if len(r) == 1 {
		return textToken(r[:1]), 1, nil
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
		if !array && r[idx] == '(' {
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
			return nil, 0, fmt.Errorf("missing close-brace for variable name")
		}
		if array {
			return nil, 0, fmt.Errorf("missing )")
		}
	}

	if !array {
		name = r[start:idx]
		if len(name) == 0 {
			return textToken(r[:1]), 1, nil
		}
	} else {
		index = r[start:idx]
	}

	if brace {
		idx++
	}
	if array {
		idx++
	}

	tok := variableToken{textToken(r[:idx]), textToken(name)}
	if index != nil {
		idxTok, _, err := ParseTokens(index, TermCloseParen, SubstAll)
		if err != nil {
			return nil, 0, err
		}
		tok = append(tok, idxTok...)
	}
	return tok, idx, nil
}

func ParseBackslashNewlineToken(r []rune) (token, int, error) {
	if len(r) < 2 || r[0] != '\\' || r[1] != '\n' {
		return nil, 0, fmt.Errorf("must start with blackslash-newline")
	}

	idx := 2

	for i := idx; i < len(r); i++ {
		c := r[i]
		if c != ' ' && c != '\t' {
			idx = i
			break
		}
	}

	return bsToken(r[:idx]), idx, nil
}

func ParseBackslashToken(r []rune) (token, int, error) {
	if len(r) < 2 || r[0] != '\\' {
		return nil, 0, nil
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
		return bsToken(r[:2]), 2, nil
	case 'b':
		return bsToken(r[:2]), 2, nil
	case 'f':
		return bsToken(r[:2]), 2, nil
	case 'n':
		return bsToken(r[:2]), 2, nil
	case 'r':
		return bsToken(r[:2]), 2, nil
	case 't':
		return bsToken(r[:2]), 2, nil
	case 'v':
		return bsToken(r[:2]), 2, nil
	case '\\':
		return bsToken(r[:2]), 2, nil
	case '\n':
		return ParseBackslashNewlineToken(r)
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
		return bsToken(r[:len(buf)+idx]), len(buf) + idx, nil
	}

	// Backslash substitution is not performed on words enclosed
	// in braces, except for backslash-newline as described above.

	// In all cases but those described below the backslash is
	// dropped and the following character is treated as an
	// ordinary character and included in the word.
	return bsToken(r[:2]), 2, nil
}

func SubstBackslashToken(r bsToken) (token, error) {
	if len(r) < 2 || r[0] != '\\' {
		return nil, fmt.Errorf("must be characters and start with backslash")
	}

	first := r[1]

	switch first {
	case 'a':
		return textToken([]rune{'\a'}), nil
	case 'b':
		return textToken([]rune{'\b'}), nil
	case 'f':
		return textToken([]rune{'\f'}), nil
	case 'n':
		return textToken([]rune{'\n'}), nil
	case 'r':
		return textToken([]rune{'\r'}), nil
	case 't':
		return textToken([]rune{'\t'}), nil
	case 'v':
		return textToken([]rune{'\v'}), nil
	case '\\':
		return textToken([]rune{'\\'}), nil
	case '\n':
		return textToken([]rune{' '}), nil
	}

	var (
		idx  int
		base int
	)

	switch first {
	case '0', '1', '2', '3',
		'4', '5', '6', '7':
		idx = 1
		base = 8
	case 'x', 'u', 'U':
		idx = 2
		base = 16
	default:
		return textToken([]rune{r[1]}), nil
	}

	val, err := strconv.ParseInt(string(r[idx:]), base, 64)
	if err != nil {
		return nil, err
	}

	return textToken([]rune{rune(val)}), nil
}

var (
	_ word = wordToken{}
	_ word = simpleWordToken{}
	_ word = expandWordToken{}

	_ token = textToken{}
	_ token = bsToken{}
	_ token = commandToken{}
	_ token = variableToken{}

	_ token = subExprToken{}
	_ token = operatorToken{}
)

type word interface {
	fmt.Stringer
	isWord()
	Subst(substs SubstType) (string, error)
}

type words []word

func (ws words) String() string {
	var b strings.Builder
	for i := 0; i < len(ws); i++ {
		_, err := b.WriteString(ws[i].String())
		if err != nil {
			return ""
		}
	}
	return b.String()
}

func (ws words) Subst(substs SubstType) (string, error) {
	var b strings.Builder
	for i := 0; i < len(ws); i++ {
		w := ws[i]
		s, err := w.Subst(substs)
		if err != nil {
			return "", err
		}
		_, err = b.WriteString(s)
		if err != nil {
			return "", err
		}

	}
	return b.String(), nil
}

// This token ordinarily describes one word of a command but it may
// also describe a quoted or braced string in an expression.
type wordToken tokens

func (_ wordToken) isWord() {}

func (w wordToken) String() string {
	var b strings.Builder
	for i := 0; i < len(w); i++ {
		_, err := b.WriteString(w[i].String())
		if err != nil {
			return ""
		}
	}
	return b.String()
}

func (w wordToken) Subst(substs SubstType) (string, error) {
	var b strings.Builder
	for i := 0; i < len(w); i++ {
		tok := w[i]
		s, err := tok.Subst(substs)
		if err != nil {
			return "", err
		}
		_, err = b.WriteString(s)
		if err != nil {
			return "", err
		}

	}
	return b.String(), nil
}

// This token has the same meaning as wordToken, except that the word
// is guaranteed to consist of a single textToken sub-token.
type simpleWordToken textToken

func (_ simpleWordToken) isWord() {}

func (w simpleWordToken) String() string { return string(w) }

func (w simpleWordToken) Subst(substs SubstType) (string, error) {
	s, err := textToken(w).Subst(substs)
	if err != nil {
		return "", err
	}
	return s, nil
}

// This token has the same meaning as wordToken, except that the
// command parser notes this word began with the expansion prefix {*}.
type expandWordToken struct {
	word
}

func (w expandWordToken) String() string { return "{*}" + w.word.String() }

type token interface {
	fmt.Stringer
	isToken()
	Subst(substs SubstType) (string, error)
}

type tokens []token

func (ts tokens) String() string {
	var b strings.Builder
	for i := 0; i < len(ts); i++ {
		_, err := b.WriteString(ts[i].String())
		if err != nil {
			return ""
		}
	}
	return b.String()
}

func (ts tokens) Subst(substs SubstType) (string, error) {
	var b strings.Builder
	for i := 0; i < len(ts); i++ {
		tok := ts[i]
		s, err := tok.Subst(substs)
		if err != nil {
			return "", err
		}
		_, err = b.WriteString(s)
		if err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

// The token describes a range of literal text that is part of a word.
type textToken []rune

func (t textToken) isToken() {}

func (t textToken) String() string { return string(t) }

func (t textToken) Subst(substs SubstType) (string, error) { return t.String(), nil }

// The token describes a backslash sequence such as \n or \0xa3.
type bsToken textToken

func (t bsToken) isToken() {}

func (t bsToken) String() string { return string(t) }

func (t bsToken) Subst(substs SubstType) (string, error) {
	tok, err := SubstBackslashToken(t)
	if err != nil {
		return "", err
	}
	return tok.String(), nil
}

// The token describes a command whose result must be substituted into
// the word.
type commandToken textToken

func (t commandToken) isToken() {}

func (t commandToken) String() string { return string(t) }

func (t commandToken) Subst(substs SubstType) (string, error) { return t.String(), nil }

// The token describes a variable substitution, including the $,
// variable name, and array index (if there is one) up through the
// close parenthesis that terminates the index.
type variableToken []token

func (t variableToken) isToken() {}

func (t variableToken) String() string {
	if len(t) == 0 {
		return ""
	}
	return t[0].String()
}

func (t variableToken) Subst(substs SubstType) (string, error) { return t.String(), nil }

// The token describes one subexpression of an expression (or an
// entire expression).
type subExprToken wordToken

func (t subExprToken) isToken() {}

func (t subExprToken) String() string { return wordToken(t).String() }

func (t subExprToken) Subst(substs SubstType) (string, error) { return t.String(), nil }

// The token describes one operator of an expression such as && or
// hypot.
type operatorToken textToken

func (t operatorToken) isToken() {}

func (t operatorToken) String() string { return textToken(t).String() }

func (t operatorToken) Subst(substs SubstType) (string, error) { return t.String(), nil }
