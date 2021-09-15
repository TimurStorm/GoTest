package urls

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
)

func GetUrls(filename string) map[int]string {
	// Для урлов
	var urls = make(map[int]string)

	// Открываем файл
	file, err := os.Open(filename + ".txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	var count = 0

	// Инициализируем сканер и считываем построчно
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		urls[count] = string(line)
		count++
	}

	return urls
}

func SendRequest(url string) *http.Response {
	// Отправляем запрос
	responce, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	return responce
}
