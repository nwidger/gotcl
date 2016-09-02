package gotcl

import "testing"

type backslashSubstTest struct {
	b          []rune
	expected   []rune
	expectedOk bool
}

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

func TestBackslashNewlineSubstOnce(t *testing.T) {
	for _, btest := range []backslashSubstTest{
		{[]rune("\\\n"), []rune(" "), true},
		{[]rune("\\\n  	hello"), []rune(" hello"), true},
	} {
		actual, actualOk := backslashNewlineSubstOnce(btest.b)
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
		actual, actualOk := backslashSubstOnce(btest.b)
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

type varSubstTest struct {
	b           []rune
	v           variableFunc
	expected    []rune
	expectedIdx int
	expectedErr error
}

func TestVarSubstOnce(t *testing.T) {
	for _, btest := range []varSubstTest{
		{[]rune(`${hello}  `), noopVariableFunc, []rune("hello  "), 5, nil},
	} {
		actual, actualIdx, actualErr := VarSubstOnce(btest.b, btest.v)
		if !runeSlicesEqual(btest.expected, actual) {
			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
		}
		if actualIdx != btest.expectedIdx {
			t.Fatalf("expected idx %v got %v for %q", btest.expectedIdx, actualIdx, btest.b)
		}
		if actualErr != btest.expectedErr {
			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
		}
	}
}
