package word

import "strings"

func IsPalindrome(word string) bool {
	normalizedRunes := []rune(strings.ToLower(word))
	var i, j int
	// double pointers, one at start one at the end of string
	for i = range len(normalizedRunes) / 2 {
		j = len(normalizedRunes) - 1 - i
		if normalizedRunes[i] != normalizedRunes[j] {
			return false
		}
	}
	return true
}
