package gotcl

import (
	"fmt"
	"strings"
)

type Interp struct {
	vars map[string]string
}

func NewInterp() *Interp {
	return &Interp{
		vars: make(map[string]string),
	}
}

func (interp *Interp) GetVar(varName string) (string, error) {
	value, ok := interp.vars[varName]
	if !ok {
		return "", fmt.Errorf(`can't read "%s": no such variable`, varName)
	}
	return value, nil
}

func (interp *Interp) GetVar2(name, index string) (string, error) {
	if len(index) == 0 {
		return interp.GetVar(name)
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
	var result string
	r := []rune(script)
	for idx := 0; idx < len(r); idx++ {
		ts, size, err := ParseCommand(r[idx:], false)
		if err != nil {
			return "", err
		}
		if size > 0 {
			idx += size - 1
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
		result = ""
		if len(ws) == 0 {
			continue
		}
		switch ws[0] {
		case "puts":
			if len(ws) > 1 {
				fmt.Println(ws[1])
			}
		case "set":
			if len(ws) == 2 {
				name := ws[1]
				return interp.GetVar(name)
			}
			if len(ws) > 2 {
				name, value := ws[1], ws[2]
				interp.vars[name] = value
				result = value
			}
		default:
			result = fmt.Sprintf("%q", ws)
		}
	}
	return result, nil
}
