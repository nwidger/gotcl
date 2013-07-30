// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:23:26 by nwidger on macros.local>

package gotcl

import (
	"fmt"
	"os"
	"strconv"
)

type Value string

func (value Value) Len() int {
	return len(string(value))
}

func (value Value) String() string {
	return string(value)
}

func (value Value) Int() int {
	i, err := strconv.Atoi(value.String())

	if err != nil {
		fmt.Printf("Cannot convert \"%s\" to an integer\n", value.String())
		os.Exit(1)
	}

	return i
}

func StringToValue(str string) Value {
	return Value(str)
}

func StringToValueP(str string) *Value {
	value := Value(str)
	return &value
}

func IntToValue(i int) Value {
	str := strconv.FormatInt(int64(i), 10)
	return Value(str)
}

func IntToValueP(i int) *Value {
	str := strconv.FormatInt(int64(i), 10)
	value := Value(str)
	return &value
}
