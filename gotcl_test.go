package gotcl

import (
	"fmt"
	"testing"
)

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

// type backslashSubstTest struct {
// 	b          []rune
// 	expected   []rune
// 	expectedOk bool
// }

// func TestBackslashNewlineSubstOnce(t *testing.T) {
// 	for _, btest := range []backslashSubstTest{
// 		{[]rune("\\\n"), []rune(" "), true},
// 		{[]rune("\\\n  	hello"), []rune(" hello"), true},
// 	} {
// 		actual, actualOk := BackslashNewlineSubstOnce(btest.b)
// 		if actualOk != btest.expectedOk {
// 			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
// 		}
// 		if !runeSlicesEqual(btest.expected, actual) {
// 			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		}
// 	}
// }

// func TestBackslashNewlineSubst(t *testing.T) {
// 	for _, btest := range []backslashSubstTest{
// 		{[]rune("\\\n"), []rune(" "), true},
// 		{[]rune("\\\n  	hello"), []rune(" hello"), true},
// 		{[]rune("\\\n  	hello  	\\\n   bye"), []rune(" hello  	 bye"), true},
// 	} {
// 		actual, actualOk := BackslashNewlineSubst(btest.b)
// 		if actualOk != btest.expectedOk {
// 			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
// 		}
// 		if !runeSlicesEqual(btest.expected, actual) {
// 			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		}
// 	}
// }

// func TestBackslashSubstOnce(t *testing.T) {
// 	for _, btest := range []backslashSubstTest{
// 		{[]rune(`\n`), []rune("\n"), true},
// 		{[]rune(`\na`), []rune("\na"), true},
// 		{[]rune(`\naaa`), []rune("\naaa"), true},
// 		{[]rune(`\001`), []rune("\001"), true},
// 		{[]rune(`\400`), []rune("\0400"), true},
// 		{[]rune(`\xfe`), []rune("þ"), true},
// 		{[]rune(`\xff`), []rune("ÿ"), true},
// 	} {
// 		actual, actualOk := BackslashSubstOnce(btest.b)
// 		if actualOk != btest.expectedOk {
// 			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
// 		}
// 		if !runeSlicesEqual(btest.expected, actual) {
// 			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		}
// 	}
// }

// func TestBackslashSubst(t *testing.T) {
// 	for _, btest := range []backslashSubstTest{
// 		{[]rune(`hellobye`), []rune("hellobye"), false},
// 		{[]rune(`hello\nbye\nhello\nbye`), []rune("hello\nbye\nhello\nbye"), true},
// 		{[]rune(`hello\n\x0abye`), []rune("hello\n\nbye"), true},
// 	} {
// 		actual, actualOk := BackslashSubst(btest.b)
// 		if actualOk != btest.expectedOk {
// 			t.Fatalf("expected %v got %v for %q", btest.expectedOk, actualOk, btest.b)
// 		}
// 		if !runeSlicesEqual(btest.expected, actual) {
// 			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		}
// 	}
// }

// func noopVariableFunc(name, index []rune) (value []rune, err error) {
// 	value = []rune{}
// 	value = append(value, name...)
// 	if index != nil {
// 		value = append(value, '(')
// 		value = append(value, index...)
// 		value = append(value, ')')
// 	}
// 	return value, nil
// }

// type parseWordTest struct {
// 	b            []rune
// 	nested       bool
// 	expected     word
// 	expectedSize int
// 	expectedErr  error
// }

// func TestParseWord(t *testing.T) {
// 	for _, btest := range []parseWordTest{
// 		{[]rune(`hello[ set x ]$arr goodbye`), false, nil, 18, nil},
// 		{[]rune(`hello$arr(\xfffin${de}x$keep)  goodbye  `), false, nil, 18, nil},
// 	} {
// 		t.Logf("--------------------------------------------------------------------------------")
// 		actual, actualSize, actualErr := ParseWord(btest.b, btest.nested)
// 		t.Logf("%q %v %v", actual, actualSize, actualErr)
// 		actualWord, actualErr := ParseWordToken(actual, btest.nested)
// 		t.Logf("%q %v", actualWord, actualErr)
// 		// if !runeSlicesEqual(btest.expected, actual) {
// 		// 	t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		// }
// 		// if actualSize != btest.expectedSize {
// 		// 	t.Fatalf("expected size %v got %v for %q", btest.expectedSize, actualSize, btest.b)
// 		// }
// 		// if actualErr != btest.expectedErr {
// 		// 	t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
// 		// }
// 	}
// }

// type parseDoubleQuoteWordTest struct {
// 	b            []rune
// 	nested       bool
// 	expected     word
// 	expectedSize int
// 	expectedErr  error
// }

// func TestParseDoubleQuoteWord(t *testing.T) {
// 	for _, btest := range []parseDoubleQuoteWordTest{
// 		{[]rune(`"hello"`), false, nil, 7, nil},
// 		{[]rune(`"hello" bye`), false, nil, 7, nil},
// 		{[]rune(`"hello
// hi" bye`), false, nil, 7, nil},
// 		{[]rune(`"hello \" goodbye" bye`), false, nil, 18, nil},
// 		{[]rune(`"hello \" goodbye" bye`), false, nil, 18, nil},
// 		{[]rune(`"hello \" $goodbye [set x]"`), false, nil, 18, nil},
// 		{[]rune(`"hello \" $goodbye [set x] "`), false, nil, 18, nil},
// 		{[]rune(`"hello \" $goodbye [set x]  "`), false, nil, 18, nil},
// 		{[]rune(`"hello \" $goodbye [ set x  ]  "`), false, nil, 18, nil},
// 		{[]rune(`"hello $arr(idx)  goodbye  "`), false, nil, 18, nil},
// 		{[]rune(`"hello $arr(in${de}x)  goodbye  "`), false, nil, 18, nil},
// 		{[]rune(`"hello $arr(\xfffin${de}x$keep)  goodbye  "`), false, nil, 18, nil},
// 	} {
// 		t.Logf("--------------------------------------------------------------------------------")
// 		actual, actualSize, actualErr := ParseQuotedStringWord(btest.b, btest.nested)
// 		t.Logf("%q %v %v", actual, actualSize, actualErr)
// 		actualWord, actualErr := ParseQuotedStringToken(actual, btest.nested)
// 		t.Logf("%q %v", actualWord, actualErr)
// 		// if !runeSlicesEqual(btest.expected, actual) {
// 		// 	t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		// }
// 		// if actualSize != btest.expectedSize {
// 		// 	t.Fatalf("expected size %v got %v for %q", btest.expectedSize, actualSize, btest.b)
// 		// }
// 		// if actualErr != btest.expectedErr {
// 		// 	t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
// 		// }
// 	}
// }

// type parseBraceWordTest struct {
// 	b            []rune
// 	nested       bool
// 	expected     word
// 	expectedSize int
// 	expectedErr  error
// }

// func TestParseBraceWord(t *testing.T) {
// 	for _, btest := range []parseBraceWordTest{
// 		{[]rune(`{hello}`), false, nil, 7, nil},
// 		{[]rune(`{hello \
//                  goodbye}`), false, nil, 18, nil},
// 		{[]rune(`{hello \
// }`), false, nil, 18, nil},
// 		{[]rune(`{hello \" $goodbye [set x]}`), false, nil, 18, nil},
// 		{[]rune(`{hello \" $goodbye [set x] }`), false, nil, 18, nil},
// 		{[]rune(`{hello \" $goodbye [set x]  }`), false, nil, 18, nil},
// 	} {
// 		t.Logf("--------------------------------------------------------------------------------")
// 		actual, actualSize, actualErr := ParseBracesWord(btest.b, btest.nested)
// 		t.Logf("%q %v %v", actual, actualSize, actualErr)
// 		actualWord, actualErr := ParseBracesToken(actual, btest.nested)
// 		t.Logf("%q %v", actualWord, actualErr)

// 	}
// }

// type parseWordsTest struct {
// 	b           []rune
// 	expected    [][]rune
// 	expectedErr error
// }

// func TestParseWords(t *testing.T) {
// 	for _, btest := range []parseWordsTest{
// 		{[]rune(`  {hello}`), [][]rune{[]rune("hello")}, nil},
// 		{[]rune(`  {hello}  "bye" what `), [][]rune{[]rune("hello"), []rune("bye"), []rune("what")}, nil},
// 		{[]rune(`puts  \"hello `), [][]rune{[]rune("puts"), []rune("\\\"hello")}, nil},
// 		{[]rune(`  {set x}`), [][]rune{[]rune("set x")}, nil},
// 		{[]rune(`  hi[ set x ]`), [][]rune{[]rune("hi[ set x ]")}, nil},
// 		{[]rune(`  "hi [ set x ] bye"`), [][]rune{[]rune("hi [ set x ] bye")}, nil},
// 		{[]rune(`  "hi ] bye"`), [][]rune{[]rune("hi ] bye")}, nil},
// 		{[]rune(`  "hi ; bye"`), [][]rune{[]rune("hi ; bye")}, nil},
// 		{[]rune(`  {*}{hi bye}`), [][]rune{[]rune("hi bye")}, nil},
// 		{[]rune(`if {x} {puts "hello";}`), [][]rune{[]rune("if"), []rune("x"), []rune("puts \"hello\";")}, nil},
// 		{[]rune(`if "x" {puts "hello";}`), [][]rune{[]rune("if"), []rune("x"), []rune("puts \"hello\";")}, nil},
// 	} {
// 		actual, _, actualErr := ParseWords(btest.b)
// 		if !runeSliceOfSlicesEqual(btest.expected, actual) {
// 			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		}
// 		if actualErr != btest.expectedErr {
// 			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
// 		}
// 	}
// }

// type parseCommandTest struct {
// 	b           []rune
// 	expected    [][]rune
// 	expectedErr error
// }

// func TestParseCommand(t *testing.T) {
// 	for _, btest := range []parseCommandTest{
// 		{[]rune(`  {hello}`), [][]rune{[]rune("hello")}, nil},
// 		{[]rune(`  {hello}  "bye" what `), [][]rune{[]rune("hello"), []rune("bye"), []rune("what")}, nil},
// 		{[]rune(`puts  \"hello `), [][]rune{[]rune("puts"), []rune("\"hello")}, nil},
// 		{[]rune(`  {set x}`), [][]rune{[]rune("set x")}, nil},
// 		{[]rune(`  "hi ] bye"`), [][]rune{[]rune("hi ] bye")}, nil},
// 		{[]rune(`  "hi ; bye"`), [][]rune{[]rune("hi ; bye")}, nil},
// 		{[]rune(`  {*}{hi bye}`), [][]rune{[]rune("hi bye")}, nil},
// 		{[]rune(`if {x} {puts "hello";}`), [][]rune{[]rune("if"), []rune("x"), []rune("puts \"hello\";")}, nil},
// 		{[]rune(`if "x" {puts "hello";}`), [][]rune{[]rune("if"), []rune("x"), []rune("puts \"hello\";")}, nil},
// 	} {
// 		actual, _, actualErr := ParseCommand(btest.b)
// 		if !runeSliceOfSlicesEqual(btest.expected, actual) {
// 			t.Fatalf("expected %q got %q for %q", btest.expected, actual, btest.b)
// 		}
// 		if actualErr != btest.expectedErr {
// 			t.Fatalf("expected err %v got %v for %q", btest.expectedErr, actualErr, btest.b)
// 		}
// 	}
// }

func dumpWords(ws words) {
	for _, x := range ws {
		fmt.Println("======================================================================")
		switch w := x.(type) {
		case wordToken:
			fmt.Printf("wordToken %q\n", w)
			for _, y := range w {
				switch t := y.(type) {
				case textToken:
					fmt.Printf("  textToken %q\n", t)
				case bsToken:
					fmt.Printf("  bsToken %q\n", t)
				case commandToken:
					fmt.Printf("  commandToken %q\n", t)
				case variableToken:
					fmt.Printf("  variableToken %q\n", t)
				case subExprToken:
					fmt.Printf("  subExprToken %q\n", t)
				case operatorToken:
					fmt.Printf("  operatorToken %q\n", t)
				}
			}
		case simpleWordToken:
			fmt.Printf("simpleWordToken %q\n", w)
			fmt.Printf("  textToken %q\n", textToken(w))
		case expandWordToken:
			fmt.Printf("expandWordToken %q\n", w)
			switch u := w.word.(type) {
			case wordToken:
				fmt.Printf("  wordToken %q\n", u)
				for _, y := range u {
					switch t := y.(type) {
					case textToken:
						fmt.Printf("    textToken %q\n", t)
					case bsToken:
						fmt.Printf("    bsToken %q\n", t)
					case commandToken:
						fmt.Printf("    commandToken %q\n", t)
					case variableToken:
						fmt.Printf("    variableToken %q\n", t)
					case subExprToken:
						fmt.Printf("    subExprToken %q\n", t)
					case operatorToken:
						fmt.Printf("    operatorToken %q\n", t)
					}
				}
			case simpleWordToken:
				fmt.Printf("  simpleWordToken %q\n", u)
				fmt.Printf("    textToken %q\n", textToken(u))
			}
		}
	}
}

func TestParseCommand(t *testing.T) {
	ws, size, err := ParseCommand([]rune(`set s(name)\
 {*}"hi $x bye" \
    xyz[set y 2]123
`), false)
	fmt.Println(size, err)
	dumpWords(ws)

	ws, size, err = ParseCommand([]rune(`if\nelse {$x} {
  return \
    [calc $x]
}
`), false)
	fmt.Println(size, err)
	dumpWords(ws)

	ws, size, err = ParseCommand([]rune(`puts $s(stackName) hi\;there`), false)
	fmt.Println(size, err)
	dumpWords(ws)

}
