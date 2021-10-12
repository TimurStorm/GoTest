package words

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

var simpleFiles = [][]string{
	{"url_test.txt", "result_test.json", "10"},
	{"url_test.txt", "result_test.json"},
}

func TestGetTopForFile(t *testing.T) {
	for index, files := range simpleFiles {
		if len(files) > 2 {
			count, err := strconv.Atoi(files[2])
			if err != nil {
				fmt.Println(err)
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

		fmt.Println(index, "ОКОНЧЕН")
		time.Sleep(time.Second)
	}

}
