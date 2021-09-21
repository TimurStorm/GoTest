package words

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func isNil(result Result) bool {
	for _, value := range result.Count {
		if value == 0 {
			return true
		}
	}
	for _, word := range result.Words {
		if word == "" {
			return true
		}
	}
	return false
}

func TestGetTopData(t *testing.T) {
	var wg sync.WaitGroup

	urlFile, urlErr := os.Open("test.txt")
	if urlErr != nil {
		fmt.Println("Ошибка чтения файла: ", urlErr)
	} else {
		scanner := bufio.NewScanner(urlFile)
		data := []Result{}
		for scanner.Scan() {
			wg.Add(1)
			var tag string
			row := strings.Split(scanner.Text(), ",")
			url := row[0]
			// Если указан необходимый тег в файле
			if len(row) == 2 {
				tag = row[1]
			}
			GetTopData(url, &data, &wg, tag)
		}

		for _, result := range data {
			if isNil(result) {
				t.Error("Nil element")
			}
		}
	}
}
