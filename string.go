package main

import (
	"regexp"
	"strings"
)

// Transforms any adjacent whitespace into a single space
// and removes any leading or trailing whitespace.
func trim(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(str), " ")
}

func spaces(count int) string {
	str := ""
	for i := 0; i < count; i++ {
		str += " "
	}
	return str
}
