package gotcl

func Subst(r []rune, variable variableFunc) ([]rune, error) {
	nb, ok := BackslashNewlineSubst(r)
	if ok {
		r = nb
	}

	for i := 0; i < len(r); i++ {
		switch r[i] {
		case '\\':
			nb, ok = BackslashSubstOnce(r[i:])
			if ok {
				r = append(r[:i], nb...)
			}
		case '$':
			nb, length, err := VarSubstOnce(r[i:], variable)
			if err != nil {
				return nil, err
			}
			r = append(r[:i], nb...)
			i += length
		default:
			continue
		}
	}

	return r, nil
}
