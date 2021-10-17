package top3

import (
	"strings"
	"unicode"
)

// spliter разделяет строку s по разделителям из splits
func spliter(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	splitter := func(r rune) bool {
		return m[r] == 1
	}

	return strings.FieldsFunc(s, splitter)
}

// isWord проверка на слово
func isWord(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// arrayContainString проверяет на принадлежность строки к массиву
func arrayContainString(list []string, substing string) bool {
	for _, value := range list {
		if value == substing {
			return true
		}
	}
	return false
}

// upCount возвращает количество букв в строке с up case
func upCount(s string) int {
	var count = 0
	for _, r := range s {
		if unicode.IsUpper(r) {
			count++
		}
	}
	return count
}
