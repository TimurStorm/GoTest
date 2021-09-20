package result

import (
	"main/words"
	"os"
	"testing"
)

func TestPrettyWriteJSON(t *testing.T) {
	exampledata := [][]words.Result{
		{
			words.Result{
				Url:   "example",
				Words: [3]string{"ex", "am", "ple"},
				Count: [3]int{1, 2, 3}},
		},
	}
	fileJson, jsonErr := os.Create("test.json")
	if jsonErr != nil {
		t.Error("Невозможно открыть файл:", jsonErr)
	} else {
		for _, data := range exampledata {
			writeErr := PrettyWriteJSON(fileJson, data)
			if writeErr != nil {
				t.Error("Невозможно записать в файл:", writeErr)
			}
		}
	}
}
