package gotcl

func findWordStart(b []rune) ([]rune, bool) {
	var prev rune
	for i := 0; i < len(b); i++ {
		switch {
		case b[i] == '"' && prev != '\\':
		case b[i] == '{' && prev != '\\':
		case b[i] == '[' && prev != '\\':
		case b[i] == '$' && prev != '\\':
		}
		prev = b[i]
	}
	return b, false
}
