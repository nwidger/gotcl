package gotcl

func Subst(r []rune, variable variableFunc) ([]rune, error) {
	for i := 0; i < len(r); i++ {
		switch r[i] {
		case '\\':
			nb, ok := BackslashSubstOnce(r[i:])
			if ok {
				r = append(r[:i], nb...)
			}
		case '$':
			nb, length, err := ParseVar(r[i:], variable)
			if err != nil {
				return nil, err
			}
			r = append(r[:i], nb...)
			i += length - 1
		default:
			continue
		}
	}

	return r, nil
}
