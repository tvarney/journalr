package main

import (
	"unicode"
)

// WordCount is a simple word counting algorithm.
//
// This handles primarily english, though it should be usable for other
// languages which separate words like english does (e.g. spaces and
// punctuation).
func WordCount(text string) int64 {
	var count int64

	inWord := false
	for _, r := range text {
		switch {
		case unicode.IsSpace(r):
			fallthrough
		case unicode.IsPunct(r):
			fallthrough
		case unicode.IsControl(r):
			if inWord {
				count++
				inWord = false
				continue
			}
		default:
			inWord = true
		}
	}
	// If we ended on a word, count it
	if inWord {
		count++
	}

	return count
}
