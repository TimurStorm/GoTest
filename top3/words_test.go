package top3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
)

var simpleFiles = [][]string{
	{"url_test.txt", "result_test.json", "10"},
	{"url_test.txt", "result_test.json"},
}

func TestGetTopForFile(t *testing.T) {
	for _, files := range simpleFiles {
		if len(files) > 2 {
			count, err := strconv.Atoi(files[2])
			if err != nil {
				t.Error("Can't convert limit to int")
			}

			err = GetTopForFile(files[0], files[1], GetTopOptions{Tags: []string{"p", "a"}, HostReqLimit: count})
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err := GetTopForFile(files[0], files[1], GetTopOptions{Tags: []string{"p", "a"}})
			if err != nil {
				fmt.Println(err)
			}
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

type TopWords struct {
	text  string
	words [3]string
	count [3]int
}

var simpleTexts = []TopWords{
	{
		text:  "Задача пример задачи задач примера результат",
		words: [3]string{"задач", "пример", "результат"},
		count: [3]int{3, 2, 1},
	},
	{
		text:  "Была получена грамота, а нужно было много грамот. Если Грамот не одумается, то полученая ситуация принесёт нам много проблем",
		words: [3]string{"Грамот", "получена", "много"},
		count: [3]int{3, 2, 2},
	},
}

func TestGetTopWords(t *testing.T) {
	for _, top := range simpleTexts {
		words, count, err := GetTopWords(top.text)
		if err != nil {
			t.Error("Error GetTopWords")
		}
		if words != top.words {
			t.Error("Untrust words: ", words)
		}
		if count != top.count {
			t.Error("Untrust count: ", count)
		}
	}
}
