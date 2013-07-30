// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:31:44 by nwidger on macros.local>

package gotcl

import (
	"fmt"
	"io/ioutil"
	"os"
)

func usage() {
	fmt.Printf("%s: FILE\n", os.Args[0])
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	bytes, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		fmt.Println("Error opening file %s: %s\n", os.Args[1], err)
		os.Exit(1)
	}

	script := string(bytes)
	interp := NewInterp()
	interp.Eval(script)
}
