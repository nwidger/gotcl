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

func dumpTokens(ts tokens) {
	for _, x := range ts {
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
			switch u := w.token.(type) {
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
		`$xyz`, `$::abc::xyz`, `$::abc::xyz(asdf)`, `${::abc::xyz(asdf)}`, `$xyz(asdf$xyz%)`, `${xyz(asdf)}`, `${xyz(asdf))}`, ` set s(name)\
 {*}"hi $x ${x(hi)} bye" \
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
# blah blah
# blah blah blah
# blah
set [set $name]($index) $::best::friend $ {*}[ my_cool_proc arg1 arg2 ]s`,
	} {
		ws, size, err := ParseCommand([]rune(str), false)
		fmt.Println(size, err)
		dumpTokens(ws)
	}
}

func TestParseComment(t *testing.T) {
	for _, str := range []string{`# hello there`, `
# hello there\
 friend
set x 123`, `
# hello there
# hello there
# hello there

set x 123`} {
		r, size, err := ParseComment([]rune(str))
		fmt.Printf("%v %v %v %q\n", string(r), size, err, str[size:])
	}
}

func TestParseCommands(t *testing.T) {
	for _, str := range []string{
		`
# blah blah blah
# blah blah
set s(name)\
 {*}"hi $x bye" \
    xyz[set y 2]123

# blah blah blah
# blah blah
puts hello

# blah blee bloo
if {1} {
  puts \
      goodbye
}
`,
	} {
		r := []rune(str)
		for idx := 0; idx < len(r); idx++ {
			ws, size, err := ParseCommand(r[idx:], false)
			if err != nil {
				fmt.Println(err)
				break
			}
			if size > 0 {
				idx += size - 1
			}
			fmt.Println(size, err)
			dumpTokens(ws)
		}
	}
}

func TestSubstTokens(t *testing.T) {
	for _, str := range []string{
		`
# blah blah blah
# blah blah
puts "stdout hello $there \
           to you too" { this is \
                         a brace token }
`,
	} {
		fmt.Println("======================================================================")
		ts, size, err := ParseCommand([]rune(str), false)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(size, err)
		dumpTokens(ts)
		for i := 0; i < len(ts); i++ {
			s, err := SubstTokens(nil, SubstAll, ts[i])
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("%q\n", s)
		}
	}
}

func TestEval(t *testing.T) {
	interp := NewInterp()

	for _, str := range []string{
		`
# blah blah blah
# blah blah
puts "stdout [ blah blah blah ] hello $there(is${a}way$out) \
           to you too" { this is \
                         a brace token }
`,
	} {
		fmt.Println("======================================================================")
		fmt.Println(str)
		s, err := interp.Eval(str)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(s, err)
	}
}
