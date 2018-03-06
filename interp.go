package gotcl

import (
	"fmt"
	"strings"
)

type Interp struct {
}

func NewInterp() *Interp {
	return &Interp{}
}

func (interp *Interp) GetVar(name, index string) (string, error) {
	if len(index) == 0 {
		return name, nil
	}
	var (
		b   strings.Builder
		err error
	)
	_, err = b.WriteString(name)
	_, err = b.WriteString("(")
	_, err = b.WriteString(index)
	_, err = b.WriteString(")")
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (interp *Interp) EvalTokens(tok token) (string, error) {
	return SubstTokens(interp, SubstAll, tok)
}

func (interp *Interp) Eval(script string) (string, error) {
	ts, _, err := ParseCommand([]rune(script), false)
	if err != nil {
		return "", err
	}
	ws := make([]string, 0, len(ts))
	for i := 0; i < len(ts); i++ {
		tok := ts[i]
		s, err := SubstTokens(interp, SubstAll, tok)
		if err != nil {
			return "", err
		}
		ws = append(ws, s)
	}
	return fmt.Sprintf("%q", ws), nil
}
