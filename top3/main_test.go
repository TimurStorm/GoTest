package top3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

type TestText struct {
	Url   string
	Tags  []string
	Text  string
	Words [3]string
	Count [3]int
}

var TestUrls = []TestText{
	{
		Url:   "https://www.rfc-editor.org/rfc/rfc172.txt",
		Words: [3]string{"file", "data", "transfer"},
		Count: [3]int{86, 58, 47},
	},
	{
		Url: "https://fabulae.ru/prose_b.php?id=75081",
		Tags: []string{
			"h1",
		},
		Text:  "Собака Сталина",
		Words: [3]string{"собак", "собака", "миша"},
		Count: [3]int{34, 19, 18},
	},
}

func TestExtractText(t *testing.T) {
	for index, test := range TestUrls {
		resp, err := http.Get(test.Url)
		if err != nil {
			t.Errorf("Ошибка отправки запроса: %v для %v", err, test.Url)
		}

		respByte, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			t.Errorf("Ошибка конвертации в байты: %v для %v", err, test.Url)
		}

		text, err := extractText(respByte, test.Tags...)
		if err != nil {
			t.Errorf("Ошибка получения текста: %v для %v", err, test.Url)
		}

		if text != test.Text && index > 0 && len(text) != 0 {
			t.Errorf("Получен неверный текст для %v", test.Url)
		}
	}
}

type TestWords struct {
	Text  string
	Words [3]string
	Count [3]int
}

var TestTex = []TestWords{
	{
		Text: "Хорош хороша хороша хороши тест тестировать тесты зло",
		Words: [3]string{
			"хорош", "тест", "хороша",
		},
		Count: [3]int{
			4, 3, 2,
		},
	},
	{
		Text: "Собака Сталина",
		Words: [3]string{
			"собака", "сталина", "",
		},
		Count: [3]int{
			1, 1, 0,
		},
	},
}

func TestGetPopularWords(t *testing.T) {
	for index, test := range TestTex {
		words, count, err := GetPopularWords(test.Text)
		if err != nil {
			fmt.Println(err)
		}
		if words != test.Words {
			t.Errorf("Для теста %v ожидались слова %v было получено %v", index, test.Words, words)
		}
		if count != test.Count {
			t.Errorf("Для теста %v ожидались значения %v было получено %v", index, test.Count, count)
		}
	}
}

func TestGetTop(t *testing.T) {
	for _, test := range TestUrls {
		top, err := GetTop(test.Url)
		if err != nil {
			fmt.Println(err)
		}
		if top.Words != test.Words {
			t.Errorf("Для %v ожидались слова %v было получено %v", test.Url, test.Words, top.Words)
		}
		if top.Count != test.Count {
			t.Errorf("Для %v ожидались значения %v было получено %v", test.Url, test.Count, top.Count)
		}
	}
}

var simpleFiles = [][]string{
	{"url_test.txt", "result_test.json", "10"},
	{"url_test.txt", "result_test.json"},
}

func TestGetTopForFile(t *testing.T) {
	for _, files := range simpleFiles {
		err := GetTopFile(files[0], files[1])
		if err != nil {
			fmt.Println(err)
		}

		resultFile, err := os.Open("result_test.json")
		if err != nil {
			t.Error("Can't open result file")
		}

		scanner := bufio.NewScanner(resultFile)
		for scanner.Scan() {
			row := scanner.Text()
			if row == "" {
				t.Error("Null row")
			}
			err := scanner.Err()
			if err != nil {
				t.Error("File error")
				break
			}
			var result Result
			err = json.Unmarshal([]byte(row), &result)
			if err != nil {
				t.Error("Unmarshal error")
			}
			if result.Count == [3]int{0, 0, 0} || result.Words == [3]string{"", "", ""} {
				t.Error("Untrust result")
			}
		}
	}
}
