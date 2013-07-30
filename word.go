// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:22:34 by nwidger on macros.local>

package gotcl

type Word interface {
	String() string
}

type LiteralWord string

func (word LiteralWord) String() string { return string(word) }

type BraceWord string

func (word BraceWord) String() string { return string(word) }

type DoubleQuoteWord string

func (word DoubleQuoteWord) String() string { return string(word) }
