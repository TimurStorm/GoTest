package top3

import (
	"fmt"
	"testing"
)

var files = []string{
	"url_test.txt",
	"u.txt",
}

var testRepeated = [][]string{
	{"habr.com", "fabulae.ru", "www.avito.ru"},
	{""},
}

func TestGetRepeatedHosts(t *testing.T) {
	for index, fileName := range files {
		hosts, err := getRepeatedHosts(fileName)
		if err != nil {
			fmt.Println(err)
		}
		for _, host := range hosts {
			if !arrayContainString(testRepeated[index], host) {
				t.Errorf("Хост %v не найден", host)
			}
		}
	}
}

var testUnique = [][]string{
	{"тарзан", "МГУ", "ТестВерблюжегоСтиля", "EnglishExample", "тарзан", "тАрзан"},
	{"тарзан", "мгу", "тест", "верблюжего", "стиля", "english", "example"},
}

func TestGetUnique(t *testing.T) {
	unique := getUnique(testUnique[0])
	for _, u := range unique {
		if !arrayContainString(testUnique[1], u) {
			t.Errorf("Уникальное слово %v не найдено", u)
		}
	}
}

type popularStruct struct {
	text    string
	popular map[int][]string
	max     int
}

var testPopular = popularStruct{
	text: "хороша хорошо хорош книга книг книги книгиня война сражение меч",
	popular: map[int][]string{
		4: {"книг"},
		3: {"хорош"},
		2: {"книги"},
		1: {"хороша", "хорошо", "книга", "книгиня", "война", "сражение", "меч"},
	},
	max: 4,
}

func TestGetPopularWords(t *testing.T) {
	allWords, max := getPopularWords(testPopular.text)
	if max != testPopular.max {
		t.Errorf("Max not valid")
	}
	for count, words := range allWords {
		for _, word := range words {
			if !arrayContainString(testPopular.popular[count], word) {
				t.Errorf("Word %v not contain in row %v with count %v", word, testPopular.popular[count], count)
			}
		}
	}
}
