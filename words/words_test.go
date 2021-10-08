package words

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var simpleFiles = [][]string{
	{"url_test.txt", "result_test.json", "10"},
	{"url_test.txt", "result_test.json", "0"},
	{"u_test.txt", "result_test.json", "0"},
	{"url_test.txt", "", "0"},
}

func TestGetTopForFile(t *testing.T) {
	for _, files := range simpleFiles {

		err := GetTopForFile(files[0], files[1], GetTopOptions{Tags: []string{"p", "a"}, HostReqLimit: 10})
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
