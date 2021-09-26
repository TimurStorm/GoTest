package words

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var topWordsData = []Result{
	{
		"https://habr.com/ru/post/578464/",
		[3]string{"Комментарии", "голосов", "дела"},
		[3]int{11, 10, 5},
	},
	{
		"https://testexample.com",
		[3]string{"", "", ""},
		[3]int{0, 0, 0},
	},
	{
		"https://www.rfc-editor.org/rfc/rfc172.txt",
		[3]string{"file", "data", "transfer"},
		[3]int{81, 54, 44},
	},
}

var simpleFiles = [][]string{
	{"tes.txt", "result.json"},
	{"test.txt", ""},
	{"test.txt", "result.json"},
}

func TestGetTop(t *testing.T) {
	for _, top := range topWordsData {
		result, err := GetTop(top.Url, []string{"div"})
		if err != nil {
			fmt.Println(err)
			return
		}
		if result.Words != top.Words || result.Count != top.Count {
			t.Errorf("Error: words %v with count %v, found %v with %v ", top.Words, top.Count, result.Words, result.Count)
		}
	}
}

func TestFindTopForFile(t *testing.T) {
	for _, files := range simpleFiles {
		err := FindTopForFile(files[0], files[1], []string{"div"})
		if err != nil {
			fmt.Println(err)
		}
		resultFile, err := os.Open("result.json")
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
