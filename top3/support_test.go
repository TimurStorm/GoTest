package top3

import (
	"testing"
)

type TestSupport struct {
	Word    string
	IsWord  bool
	Array   []string
	Contain bool
	UpCount int
	Spliter string
	Split   []string
}

var TestData = []TestSupport{
	{
		Word:    "Тимур",
		IsWord:  true,
		Array:   []string{"Тимур", "Тагир"},
		Contain: true,
		UpCount: 1,
		Spliter: "м",
		Split:   []string{"Ти", "ур"},
	},
	{
		Word:    "1roll",
		IsWord:  false,
		Array:   []string{"Тимур", "Тагир"},
		Contain: false,
		UpCount: 0,
		Spliter: "o",
		Split:   []string{"1r", "ll"},
	},
}

func TestIsWord(t *testing.T) {
	for _, test := range TestData {
		isWord := isWord(test.Word)
		if (isWord && !test.IsWord) || (!isWord && test.IsWord) {
			t.Errorf("Проверка на слово '%v' провалена: получено %v, ожидалось %v", test.Word, isWord, test.IsWord)
		}
	}
}

func TestArrayContainString(t *testing.T) {
	for _, test := range TestData {
		contain := arrayContainString(test.Array, test.Word)
		if (contain && !test.Contain) || (!contain && test.Contain) {
			t.Errorf("Проверка содержания слова '%v' в массиве %v провалена: получено %v, ожидалось %v",
				test.Word, test.Array, contain, test.Contain)
		}
	}
}

func TestUpCount(t *testing.T) {
	for _, test := range TestData {
		up := upCount(test.Word)
		if up != test.UpCount {
			t.Errorf("Подсчёт содержания заглавных букв в слове '%v' провален: получено %v, ожидалось %v",
				test.Word, up, test.UpCount)
		}
	}
}

func TestSpliter(t *testing.T) {
	for _, test := range TestData {
		s := spliter(test.Word, test.Spliter)
		for _, item := range s {
			if !arrayContainString(test.Split, item) {
				t.Errorf("Cлово '%v' неправильно разбито на части: получено %v, ожидалось %v",
					test.Word, s, test.Split)
				break
			}
		}
	}
}
