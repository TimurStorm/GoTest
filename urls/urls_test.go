package urls

import (
	"testing"
)

func TestGetUrls(t *testing.T) {
	all_urls := GetUrls("../test")
	for _, url := range all_urls {
		resp := SendRequest(url)
		if resp.StatusCode != 200 {
			t.Error(
				"For", url,
				"espected", 200,
				"got", resp.StatusCode,
			)
		}
	}
}
