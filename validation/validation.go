package validation

import "unicode"

func IsNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func IsNewLine(s string) bool {
	if len(s) < 1 {
		return false
	}
	return s[len(s)-1] == '\n' || (s[len(s)-1] != '\n' && s[len(s)-1] != ',')
}

func IsColumnName(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func IsOperator(r rune) bool {
	if r == '+' || r == '-' || r == '*' || r == '/' {
		return true
	}
	return false
}
