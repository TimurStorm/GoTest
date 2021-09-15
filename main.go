package main

import (
	"fmt"
	"runtime"

	"encoding/json"
	"io/ioutil"

	"main/urls"
	"main/words"
)

type Result struct {
	Url   string
	Words [3]string
	Count [3]int
}

type File struct {
	Results []Result
}

func main() {

	memStats := &runtime.MemStats{}
	// Получаем урлы
	all_urls := urls.GetUrls("url")
	file := File{}
	// Проходимся по всем урлам
	for i, url := range all_urls {

		fmt.Printf("REQUEST %x -------- \n", i)
		// Отправляем запрос
		responce := urls.SendRequest(url)
		// Получаем текст страницы
		text := words.GetText(responce, "body")

		// Получаем топ 3 упомянаемых слова с колиством упомянаний
		words, count := words.GetWordsCount(text)

		// Сохраняем полученные данные
		result := Result{Url: url, Words: words, Count: count}
		file.Results = append(file.Results, result)
	}
	// Запись в .json файл
	data, _ := json.MarshalIndent(file, "", " ")
	_ = ioutil.WriteFile("result.json", data, 0644)
	runtime.ReadMemStats(memStats)
	fmt.Printf("\nAlloc = %v\nTotalAlloc = %v\nSys = %v\nNumGC = %v\n\n", memStats.Alloc/1024, memStats.TotalAlloc/1024, memStats.Sys/1024, memStats.NumGC)
}
