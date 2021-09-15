package strings

import (
	"sort"
	"strings"
)

func FriendlyJoin(words []string) string {
	if len(words) == 0 {
		return ""
	}
	if len(words) == 1 {
		return words[0]
	}
	sort.Strings(words)
	commas := words[0 : len(words)-1]
	return strings.Join(commas, ", ") + " and " + words[len(words)-1]
}

func Pluralize(word string, count int) string {
	if count > 1 {
		return word + "s"
	}
	return word
}
