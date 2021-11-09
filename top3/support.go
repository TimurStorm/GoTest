package top3

import (
	"bufio"
	"io"
	"os"
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

// lineCounter считает количество строк в файле
func lineCounter(fileName string) (uint, error) {

	file, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var count uint

	for {

		_, isPrefix, err := reader.ReadLine()

		if !isPrefix {
			count++
		}

		if err == io.EOF {
			return count - 1, nil
		} else if err != nil {
			return count, err
		}

	}
}

// readln считывает следующую строку из файла
func readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}
