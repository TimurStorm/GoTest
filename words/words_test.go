package words

import (
	"fmt"
	"main/urls"
	"testing"
)

func TestGetWordsCount(t *testing.T) {
	all_urls := urls.GetUrls("../test")
	famous_words := [][3]string{
		{"file", "data", "transfer"},
		{"Oтветить", "меня", "Владимир"},
	}
	famous_count := [][3]int{
		{81, 54, 44},
		{19, 18, 13},
	}
	all_tags := []string{"body", "body"}
	for index, url := range all_urls {
		var no_contain bool = true
		resp := urls.SendRequest(url)
		text := GetText(resp, all_tags[index])
		words, count := GetWordsCount(text)
		fmt.Println(words, count)
		for _, word := range words {
			for i, famous := range famous_words {
				if famous[i] == word {
					no_contain = false
					break
				}
			}
			if !no_contain {
				break
			}
		}
		if famous_count[index] != count && no_contain {
			t.Error(
				"\nUrl :", url,
				"\nwords:", famous_words[index],
				"\ncount:", famous_count[index],
				"\nget_words:", words,
				"\nget_count:", count,
			)
		}
	}
}
