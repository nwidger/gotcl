package main

import (
	"testing"
)

func TestParseBraceWord(t *testing.T) {
	var ok bool
	var word string
	var script string
	var remainder string

	n := 0

	script = "not a brace word"
	ok, word, remainder = InterpParseBraceWord(script)

	t.Logf("script = ->%v<-, ok = %v, word = ->%v<-, remainder = ->%v<-\n", script, ok, word, remainder)

	if ok {
		t.Log(n, "ok failed")
		t.Fail()
	}

	if word != "" {
		t.Log(n, "word failed")
		t.Fail()
	}

	if remainder != script {
		t.Log(n, "remainder failed")
		t.Fail()
	}

	n++

	script = " { i am a brace word }"
	ok, word, remainder = InterpParseBraceWord(script)

	t.Logf("script = ->%v<-, ok = %v, word = ->%v<-, remainder = ->%v<-\n", script, ok, word, remainder)

	if !ok {
		t.Log(n, "ok failed")
		t.Fail()
	}

	if word != " i am a brace word " {
		t.Log(n, "word failed")
		t.Fail()
	}

	if remainder != "" {
		t.Log(n, "remainder failed")
		t.Fail()
	}

	n++

	script = "{ arg1 { arg2 default } }"
	ok, word, remainder = InterpParseBraceWord(script)

	t.Logf("script = ->%v<-, ok = %v, word = ->%v<-, remainder = ->%v<-\n", script, ok, word, remainder)

	if !ok {
		t.Log(n, "ok failed")
		t.Fail()
	}

	if word != " arg1 { arg2 default } " {
		t.Log(n, "word failed")
		t.Fail()
	}

	if remainder != "" {
		t.Log(n, "remainder failed")
		t.Fail()
	}

	n++
}
