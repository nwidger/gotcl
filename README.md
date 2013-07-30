GoTcl: TCL Interpreter in Go
======================================================================

GoTcl intends to be a Tcl interpreter implementing the [Tcl 8.5
syntax](http://www.tcl.tk/man/tcl8.5/TclCmd/Tcl.htm).

Commands
----------------------------------------------------------------------

Implemented.  Commands are separated by newlines or semicolons.

Evaluation
----------------------------------------------------------------------

Implemented.

Words
----------------------------------------------------------------------

Normal, double-quote, and brace words are parsed.  The {*}{a b [c] d}
word form is not implemented yet.

Command Substitution
----------------------------------------------------------------------

Command substitution is performed on words.

Variable Subsitution
----------------------------------------------------------------------

Variable substitution is performed, but looks no further than local
variables when trying to resolve a name.  The ${name} syntax has not
been implemented yet.  Arrays are not implemented yet, so obviously
the $name(elem) syntax will not work.

Backslash Substitution
----------------------------------------------------------------------

Is *not* implemented as of yet.  

Comments
----------------------------------------------------------------------

Comments should be parsed correctly.

Namespaces
----------------------------------------------------------------------

The concept of a namespace is built into the interpreter but for now
the global (::) namespace is assumed for all operations.
