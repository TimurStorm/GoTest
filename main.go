package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"main/words"
)

var wg sync.WaitGroup

func process(url string, data *[]map[string]interface{}, defaultTag ...string) {

	var tag string

	if defaultTag[0] != "" {
		tag = defaultTag[0]
	}

	fmt.Printf("REQUEST %v \n", url)
	// Отправляем запрос
	resp, respErr := http.Get(url)
	if respErr != nil {
		fmt.Println("Ошибка отправки запроса: ", respErr)
	} else {
		// Получаем текст страницы
		text, textErr := words.GetText(resp, tag)
		if textErr != nil {
			fmt.Println("Ошибка получения текста: ", textErr)
		} else {
			// Получаем топ 3 упомянаемых слова с колиством упомянаний
			words, count := words.GetWordsCount(text)

			// Сохраняем полученные данные
			result := map[string]interface{}{
				"url":   url,
				"words": words,
				"count": count,
			}
			*data = append(*data, result)
		}
	}
	wg.Done()
}

func main() {
	runtime.GOMAXPROCS(1)
	start := time.Now()

	data := []map[string]interface{}{}
	// Открываем файл
	urlFile, urlErr := os.Open("stres_test.txt")
	if urlErr != nil {
		fmt.Println("Ошибка считывания файла с url:", urlErr)
	}
	jsonFile, jsonErr := os.Create("result.json")
	if jsonErr != nil {
		fmt.Println("Ошибка считывания файла с url:", jsonErr)
	}

	defer urlFile.Close()

	// Инициализируем сканер и считываем построчно
	scanner := bufio.NewScanner(urlFile)
	count := 1
	for scanner.Scan() {
		wg.Add(1)
		var tag string
		row := strings.Split(scanner.Text(), ",")
		url := row[0]
		if len(row) == 2 {
			tag = row[1]
		}
		go process(url, &data, tag)

		count++
	}

	wg.Wait()
	encoder := json.NewEncoder(jsonFile)
	encoder.Encode(data)
	fmt.Printf("Time spend: %v", time.Since(start))

}
