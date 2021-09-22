package words

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGetTopData(t *testing.T) {
	// Канал для передачи данных
	urlResult := make(chan Result)
	urlCount := 0
	urlFile, urlErr := os.Open("test.txt")
	if urlErr != nil {
		fmt.Println("Ошибка чтения файла: ", urlErr)
	} else {
		scanner := bufio.NewScanner(urlFile)
		for scanner.Scan() {
			var tag string
			row := strings.Split(scanner.Text(), ",")
			url := row[0]
			// Если указан необходимый тег в файле
			if len(row) == 2 {
				tag = row[1]
			}
			go GetTopData(url, urlResult, tag)
			urlCount++
		}
		for ; urlCount != 0; urlCount-- {
			// Считываем результат из канала
			result := <-urlResult
			// Если нет результатов
			if result.Count == [3]int{0, 0, 0} {
				t.Error("Error")
			}
		}
		close(urlResult)
	}
}
