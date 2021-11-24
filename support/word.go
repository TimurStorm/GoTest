package support

import (
	"strings"
	"unicode"
)

// spliter разделяет строку s по всем разделителям из " -.,?!()<>_"
func Spliter(s string) []string {
	var (
		result []string
		index  int
	)
	for {
		index = strings.IndexAny(s, " -.,?!()<>_")
		if index == -1 {
			break
		}
		word := s[:index]
		if word != "" {
			result = append(result, word)
		}
		s = s[index+1:]
	}
	if s != "" {
		result = append(result, s)
	}
	return result
}

// isWord проверка на слово
func IsWord(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// arrayContainString проверяет на принадлежность строки к массиву
func ArrayContainString(list []string, substing string) bool {
	for _, value := range list {
		if value == substing {
			return true
		}
	}
	return false
}

// upCount возвращает количество букв в строке с up case
func UpCount(s string) int {
	var count = 0
	for _, r := range s {
		if unicode.IsUpper(r) {
			count++
		}
	}
	return count
}
