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
					fmt.Printf("  variableToken %q\n", []token(t))
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
						fmt.Printf("    variableToken %q\n", []token(t))
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
	for _, str := range []string{
		`
set s(name)\
 {*}"hi $x bye" \
    xyz[set y 2]123
`,
		`
if\nelse {$x} {
  return \
    [calc $x]
}
`,
		`
if {$x == 23} { \
  puts "hello $friend my $var(${name}) ${is} bob"
}
`,
		`
if {$x == 23} \
  puts "hello $friend my $var(${name}) ${is} bob"
`,
		`
set [set $name]($index) $::best::friend $ {*}[ my_cool_proc arg1 arg2 ]`,
	} {
		ws, size, err := ParseCommand([]rune(str), false)
		fmt.Println(size, err)
		dumpWords(ws)
	}
}
