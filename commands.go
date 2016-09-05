package gotcl

func ParseCommand(r []rune) ([][]rune, int, error) {
	nb, ok := BackslashNewlineSubst(r)
	if ok {
		r = nb
	}

	ws, size, err := ParseWords(r)
	if err != nil {
		return nil, 0, err
	}

	variable := func(name, index []rune) (value []rune, err error) {
		value = []rune{}
		value = append(value, name...)
		if index != nil {
			value = append(value, '(')
			value = append(value, index...)
			value = append(value, ')')
		}
		return value, nil
	}

	for i := range ws {
		ws[i], err = Subst(ws[i], variable)
		if err != nil {
			return nil, 0, err
		}
	}

	return ws, size, nil
}
