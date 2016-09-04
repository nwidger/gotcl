package gotcl

import "testing"

func runeSlicesEqual(ra, rb []rune) bool {
	if len(ra) != len(rb) {
		return false
	}
	for i := 0; i < len(ra); i++ {
		if ra[i] != rb[i] {
			return false
		}
	}
	return true
}

func runeSliceOfSlicesEqual(ra, rb [][]rune) bool {
	if len(ra) != len(rb) {
		return false
	}
	for i := 0; i < len(ra); i++ {
		if !runeSlicesEqual(ra[i], rb[i]) {
			return false
		}
	}
	return true
}

type backslashSubstTest struct {
	b          []rune
	expected   []rune
	expectedOk bool
}

func TestBackslashNewlineSubstOnce(t *testing.T) {
	for _, btest := range []backslashSubstTest{
		{[]rune("\\\n"), []rune(" "), true},
		{[]rune("\\\n  	hello"), []rune(" hello"), true},
	} {
		actual, actualOk := BackslashNewlineSubstOnce(btest.b)
		if actualOk != btest.expectedOk {
			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
		}
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
	}
}

func TestBackslashNewlineSubst(t *testing.T) {
	for _, btest := range []backslashSubstTest{
		{[]rune("\\\n"), []rune(" "), true},
		{[]rune("\\\n  	hello"), []rune(" hello"), true},
		{[]rune("\\\n  	hello  	\\\n   bye"), []rune(" hello  	 bye"), true},
	} {
		actual, actualOk := BackslashNewlineSubst(btest.b)
		if actualOk != btest.expectedOk {
			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
		}
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
	}
}

func TestBackslashSubstOnce(t *testing.T) {
	for _, btest := range []backslashSubstTest{
		{[]rune(`\n`), []rune("\n"), true},
		{[]rune(`\na`), []rune("\na"), true},
		{[]rune(`\naaa`), []rune("\naaa"), true},
		{[]rune(`\001`), []rune("\001"), true},
		{[]rune(`\400`), []rune("\0400"), true},
		{[]rune(`\xfe`), []rune("þ"), true},
		{[]rune(`\xff`), []rune("ÿ"), true},
	} {
		actual, actualOk := BackslashSubstOnce(btest.b)
		if actualOk != btest.expectedOk {
			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
		}
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
	}
}

func TestBackslashSubst(t *testing.T) {
	for _, btest := range []backslashSubstTest{
		{[]rune(`hellobye`), []rune("hellobye"), false},
		{[]rune(`hello\nbye\nhello\nbye`), []rune("hello\nbye\nhello\nbye"), true},
		{[]rune(`hello\n\x0abye`), []rune("hello\n\nbye"), true},
	} {
		actual, actualOk := BackslashSubst(btest.b)
		if actualOk != btest.expectedOk {
			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
		}
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
	}
}

func noopVariableFunc(name, index []rune) (value []rune, err error) {
	value = []rune{}
	value = append(value, name...)
	if index != nil {
		value = append(value, '(')
		value = append(value, index...)
		value = append(value, ')')
	}
	return value, nil
}

type varSubstOnceTest struct {
	b              []rune
	v              variableFunc
	expected       []rune
	expectedLength int
	expectedErr    error
}

func TestVarSubstOnce(t *testing.T) {
	for _, btest := range []varSubstOnceTest{
		{[]rune(`$hello  `), noopVariableFunc, []rune("hello  "), 5, nil},
		{[]rune(`${hello}  `), noopVariableFunc, []rune("hello  "), 5, nil},
		{[]rune(`${hello}  bye`), noopVariableFunc, []rune("hello  bye"), 5, nil},
	} {
		actual, actualLength, actualErr := VarSubstOnce(btest.b, btest.v)
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
		if actualLength != btest.expectedLength {
			t.Fatalf("expected length %v got %v for %q", btest.expectedLength, actualLength, btest.b)
		}
		if actualErr != btest.expectedErr {
			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
		}
	}
}

type varSubstTest struct {
	b           []rune
	v           variableFunc
	expected    []rune
	expectedErr error
}

func TestVarSubst(t *testing.T) {
	for _, btest := range []varSubstTest{
		//
		{[]rune(`$hello  `), noopVariableFunc, []rune("hello  "), nil},
		{[]rune(`${hello}  `), noopVariableFunc, []rune("hello  "), nil},
		{[]rune(`${hello}  bye`), noopVariableFunc, []rune("hello  bye"), nil},
		{[]rune(`${hello}  $bye`), noopVariableFunc, []rune("hello  bye"), nil},
		{[]rune(`  ${hello}  $bye `), noopVariableFunc, []rune("  hello  bye "), nil},
		//
		{[]rune(`$hello(idx)  `), noopVariableFunc, []rune("hello(idx)  "), nil},
		{[]rune(`${hello(idx)}  `), noopVariableFunc, []rune("hello(idx)  "), nil},
		{[]rune(`${hello(idx)}  bye`), noopVariableFunc, []rune("hello(idx)  bye"), nil},
		{[]rune(`${hello(idx)}  $bye(idx)`), noopVariableFunc, []rune("hello(idx)  bye(idx)"), nil},
		{[]rune(`  ${hello(idx)}  $bye(idx) `), noopVariableFunc, []rune("  hello(idx)  bye(idx) "), nil},
	} {
		actual, actualErr := VarSubst(btest.b, btest.v)
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
		if actualErr != btest.expectedErr {
			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
		}
	}
}

type substTest struct {
	b           []rune
	v           variableFunc
	expected    []rune
	expectedErr error
}

func TestSubst(t *testing.T) {
	for _, btest := range []varSubstTest{
		{[]rune(`$hello \xfe `), noopVariableFunc, []rune("hello þ "), nil},
		{[]rune(`$hello \
 \xfe `), noopVariableFunc, []rune("hello  þ "), nil},
	} {
		actual, actualErr := Subst(btest.b, btest.v)
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
		if actualErr != btest.expectedErr {
			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
		}
	}
}

type parseDoubleQuoteWordTest struct {
	b              []rune
	expected       []rune
	expectedLength int
	expectedErr    error
}

func TestParseDoubleQuoteWord(t *testing.T) {
	for _, btest := range []parseDoubleQuoteWordTest{
		{[]rune(`"hello"`), []rune("hello"), 7, nil},
		{[]rune(`"hello" bye`), []rune("hello"), 7, nil},
		{[]rune(`"hello \" goodbye" bye`), []rune("hello \\\" goodbye"), 18, nil},
	} {
		actual, actualLength, actualErr := ParseDoubleQuoteWord(btest.b)
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
		if actualLength != btest.expectedLength {
			t.Fatalf("expected length %v got %v for %q", btest.expectedLength, actualLength, btest.b)
		}
		if actualErr != btest.expectedErr {
			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
		}
	}
}

type parseBraceWordTest struct {
	b              []rune
	expected       []rune
	expectedLength int
	expectedErr    error
}

func TestParseBraceWord(t *testing.T) {
	for _, btest := range []parseBraceWordTest{
		{[]rune(`{hello}`), []rune("hello"), 7, nil},
		{[]rune(`{hello {bye\}} \{ what}bye`), []rune("hello {bye\\}} \\{ what"), 23, nil},
	} {
		actual, actualLength, actualErr := ParseBraceWord(btest.b)
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
		if actualLength != btest.expectedLength {
			t.Fatalf("expected length %v got %v for %q", btest.expectedLength, actualLength, btest.b)
		}
		if actualErr != btest.expectedErr {
			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
		}
	}
}

type parseWordsTest struct {
	b           []rune
	expected    [][]rune
	expectedErr error
}

func TestParseWords(t *testing.T) {
	for _, btest := range []parseWordsTest{
		{[]rune(`  {hello}`), [][]rune{[]rune("hello")}, nil},
		{[]rune(`  {hello}  "bye" what `), [][]rune{[]rune("hello"), []rune("bye"), []rune("what")}, nil},
		{[]rune(`puts  \"hello `), [][]rune{[]rune("puts"), []rune("\\\"hello")}, nil},
		{[]rune(`  {set x}`), [][]rune{[]rune("set x")}, nil},
		{[]rune(`  hi[ set x ]`), [][]rune{[]rune("hi[ set x ]")}, nil},
		{[]rune(`  "hi [ set x ] bye"`), [][]rune{[]rune("hi [ set x ] bye")}, nil},
		{[]rune(`  "hi ] bye"`), [][]rune{[]rune("hi ] bye")}, nil},
		{[]rune(`  "hi ; bye"`), [][]rune{[]rune("hi ; bye")}, nil},
		{[]rune(`  {*}{hi bye}`), [][]rune{[]rune("hi bye")}, nil},
	} {
		actual, actualErr := ParseWords(btest.b)
		if !runeSliceOfSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
		if actualErr != btest.expectedErr {
			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
		}
	}
}
