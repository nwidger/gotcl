package gotcl

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var (
	// commands consist of combination of wordToken,
	// simpleWordToken and expandWordToken

	// wordToken consists of combination of textToken, bsToken,
	// commandToken and variableToken.
	_ word = wordToken{}
	// simpleWordToken consists of a single textToken
	_ word = simpleWordToken{}
	// expandWordToken is a wordToken but was prefixed by
	// expansion prefix {*}
	_ word = expandWordToken{}

	// text literal
	_ token = textToken{}
	// backslash literal
	_ token = bsToken{}
	// command, including square brackets []'s
	_ token = commandToken{}
	// variable, including $, name and array index including
	// parentheses
	_ token = variableToken{}

	_ token = subExprToken{}
	_ token = operatorToken{}
)

type word interface {
	fmt.Stringer
	isWord()
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

// This token ordinarily describes one word of a command but it may
// also describe a quoted or braced string in an expression. The token
// describes a component of the script that is the result of
// concatenating together a sequence of subcomponents, each described
// by a separate subtoken. The token starts with the first non-blank
// character of the component (which may be a double-quote or open
// brace) and includes all characters in the component up to but not
// including the space, semicolon, close bracket, close quote, or
// close brace that terminates the component. The numComponents field
// counts the total number of sub-tokens that make up the word,
// including sub-tokens of TCL_TOKEN_VARIABLE and TCL_TOKEN_BS tokens.
//
// double-quote -> [textToken|commandToken|variableToken|bsToken]+
// brace        -> [textToken|bsToken (if contained backslash-newlines)]+
// otherwise    -> textToken
type wordToken []token

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

// This token has the same meaning as TCL_TOKEN_WORD, except that the
// word is guaranteed to consist of a single TCL_TOKEN_TEXT
// sub-token. The numComponents field is always 1.
type simpleWordToken textToken

func (_ simpleWordToken) isWord() {}

func (w simpleWordToken) String() string { return string(w) }

// This token has the same meaning as TCL_TOKEN_WORD, except that the
// command parser notes this word began with the expansion prefix {*},
// indicating that after substitution, the list value of this word
// should be expanded to form multiple arguments in command
// evaluation. This token type can only be created by
// Tcl_ParseCommand.
type expandWordToken struct {
	word
}

func (w expandWordToken) String() string { return "{*}" + w.word.String() }

type token interface {
	fmt.Stringer
	isToken()
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

// The token describes a range of literal text that is part of a
// word. The numComponents field is always 0.
type textToken []rune

func (t textToken) isToken() {}

func (t textToken) String() string { return string(t) }

// The token describes a backslash sequence such as \n or \0xa3. The
// numComponents field is always 0.
type bsToken textToken

func (t bsToken) isToken() {}

func (t bsToken) String() string { return string(t) }

// The token describes a command whose result must be substituted into
// the word. The token includes the square brackets that surround the
// command. The numComponents field is always 0 (the nested command is
// not parsed; call Tcl_ParseCommand recursively if you want to see
// its tokens).
//
// command (after parsing) -> [wordToken|simpleWordToken|expandWordToken]+
type commandToken textToken

func (t commandToken) isToken() {}

func (t commandToken) String() string { return string(t) }

// The token describes a variable substitution, including the $,
// variable name, and array index (if there is one) up through the
// close parenthesis that terminates the index. This token is followed
// by one or more additional tokens that describe the variable name
// and array index. If numComponents is 1 then the variable is a
// scalar and the next token is a TCL_TOKEN_TEXT token that gives the
// variable name. If numComponents is greater than 1 then the variable
// is an array: the first sub-token is a TCL_TOKEN_TEXT token giving
// the array name and the remaining sub-tokens are TCL_TOKEN_TEXT,
// TCL_TOKEN_BS, TCL_TOKEN_COMMAND, and TCL_TOKEN_VARIABLE tokens that
// must be concatenated to produce the array index. The numComponents
// field includes nested sub-tokens that are part of
// TCL_TOKEN_VARIABLE tokens in the array index.
//
// scalar -> textToken
// array  -> textToken [textToken|bsToken|commandToken|variableToken]+
type variableToken []token

func (t variableToken) isToken() {}

func (t variableToken) String() string {
	if len(t) == 0 {
		return ""
	}
	return t[0].String()
}

// The token describes one subexpression of an expression (or an
// entire expression). A subexpression may consist of a value such as
// an integer literal, variable substitution, or parenthesized
// subexpression; it may also consist of an operator and its
// operands. The token starts with the first non-blank character of

// the subexpression up to but not including the space, brace,
// close-paren, or bracket that terminates the subexpression. This
// token is followed by one or more additional tokens that describe
// the subexpression. If the first sub-token after the
// TCL_TOKEN_SUB_EXPR token is a TCL_TOKEN_OPERATOR token, the
// subexpression consists of an operator and its token operands. If
// the operator has no operands, the subexpression consists of just
// the TCL_TOKEN_OPERATOR token. Each operand is described by a
// TCL_TOKEN_SUB_EXPR token. Otherwise, the subexpression is a value
// described by one of the token types TCL_TOKEN_WORD, TCL_TOKEN_TEXT,
// TCL_TOKEN_BS, TCL_TOKEN_COMMAND, TCL_TOKEN_VARIABLE, and
// TCL_TOKEN_SUB_EXPR. The numComponents field counts the total number
// of sub-tokens that make up the subexpression; this includes the
// sub-tokens for any nested TCL_TOKEN_SUB_EXPR tokens.
type subExprToken wordToken

func (t subExprToken) isToken() {}

func (t subExprToken) String() string { return wordToken(t).String() }

// The token describes one operator of an expression such as && or
// hypot. A TCL_TOKEN_OPERATOR token is always preceded by a
// TCL_TOKEN_SUB_EXPR token that describes the operator and its
// operands; the TCL_TOKEN_SUB_EXPR token's numComponents field can be
// used to determine the number of operands. A binary operator such as
// * is followed by two TCL_TOKEN_SUB_EXPR tokens that describe its
// operands. A unary operator like - is followed by a single
// TCL_TOKEN_SUB_EXPR token for its operand. If the operator is a math
// function such as log10, the TCL_TOKEN_OPERATOR token will give its
// name and the following TCL_TOKEN_SUB_EXPR tokens will describe its
// operands; if there are no operands (as with rand), no
// TCL_TOKEN_SUB_EXPR tokens follow. There is one trinary operator, ?,
// that appears in if-then-else subexpressions such as x?y:z; in this
// case, the ? TCL_TOKEN_OPERATOR token is followed by three
// TCL_TOKEN_SUB_EXPR tokens for the operands x, y, and z. The
// numComponents field for a TCL_TOKEN_OPERATOR token is always 0.
type operatorToken textToken

func (t operatorToken) isToken() {}

func (t operatorToken) String() string { return textToken(t).String() }

type TermType int

const (
	TermQuote        TermType = 1 << iota // double-quote '"'
	TermSpace                             // whitespace
	TermCommandEnd                        // newline or semicolon
	TermCloseParen                        // close paren ')'
	TermCloseBracket                      // close bracket ']'
)

// words must only contain wordToken, simpleWordToken or
// expandWordToken
func ParseCommand(r []rune, nested bool) (words, int, error) {
	ws := words{}
	var (
		idx  int
		prev rune
	)
	for idx = 0; idx < len(r); idx++ {
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
		if prev != '\\' && (c == ';' || c == '\n' ||
			(nested && c == ']')) {
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
	if idx > len(r) {
		idx = len(r)
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
	ts, size, err := ParseTokens(r, terminators)
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

	ts, _, err := ParseTokens(r[1:len(r)-1], TermQuote)
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
func ParseArgumentExpansionWord(r []rune, nested bool) (word, int, error) {
	if len(r) < 4 ||
		r[0] != '{' ||
		r[1] != '*' ||
		r[2] != '}' ||
		unicode.IsSpace(r[3]) {
		return nil, 0, fmt.Errorf("word does not start with '{*}' followed by a non-whitespace character")
	}

	return ParseWord(r[3:], nested)
}

func ParseArgumentExpansionToken(r []rune, nested bool) (word, int, error) {
	if len(r) < 4 ||
		r[0] != '{' ||
		r[1] != '*' ||
		r[2] != '}' ||
		unicode.IsSpace(r[3]) {
		return nil, 0, fmt.Errorf("word does not start with '{*}' followed by a non-whitespace character")
	}

	nb, size, err := ParseWord(r[3:], nested)
	if err != nil {
		return nil, 0, err
	}

	return expandWordToken{nb}, size + 3, nil
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
func ParseCommandToken(r []rune) (token, int, error) {
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

	return commandToken(r[:idx+1]), idx + 1, nil
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

func ParseTokens(r []rune, terminators TermType) (tokens, int, error) {
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
			tok, size, err = ParseBackslashToken(r[idx:])
		case c == '[' && prev != '\\':
			tok, size, err = ParseCommandToken(r[idx:])
		case c == '$' && prev != '\\':
			tok, size, err = ParseVarNameToken(r[idx:])
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
	if idx > len(r) {
		idx = len(r)
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
func ParseVarNameToken(r []rune) (variableToken, int, error) {
	var (
		name  []rune
		index []rune
		idx   int
		prev  rune
	)

	if len(r) == 0 || r[0] != '$' {
		return nil, 0, fmt.Errorf("must start with '$'")
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
			return nil, 0, fmt.Errorf("missing close-brace for variable name")
		}
		if array {
			return nil, 0, fmt.Errorf("missing )")
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

	tok := variableToken{textToken(r[:idx]), textToken(name)}
	if index != nil {
		idxTok, _, err := ParseTokens(index, TermCloseParen)
		if err != nil {
			return nil, 0, err
		}
		tok = append(tok, idxTok...)
	}
	return tok, idx, nil
}
