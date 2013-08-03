// Niels Widger
// Time-stamp: <02 Aug 2013 at 21:20:34 by nwidger on macros.local>

package gotcl

type Word interface {
	String() string
	Len() int
}

type LiteralWord string

func (word LiteralWord) String() string { return string(word) }
func (word LiteralWord) Len() int       { return len(word) }

type BraceWord string

func (word BraceWord) String() string { return string(word) }
func (word BraceWord) Len() int       { return len(word) }

type DoubleQuoteWord string

func (word DoubleQuoteWord) String() string { return string(word) }
func (word DoubleQuoteWord) Len() int       { return len(word) }
