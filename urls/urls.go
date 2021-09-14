package urls

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
)

func GetUrls(filename string) map[int]string {
	var urls = make(map[int]string)
	file, err := os.Open(filename + ".txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	var count = 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		urls[count] = string(line)
		count++
	}

	return urls
}

func SendRequest(url string) *http.Response {
	responce, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	return responce
}
