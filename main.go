package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"main/result"
	"main/words"
)

func main() {
	// Используем 1 ядро процессора
	runtime.GOMAXPROCS(1)
	// Для измерения времени
	start := time.Now()
	// Количество горутин
	urlCount := 0
	// Массив результирующих данных
	var data []words.Result
	// Канал для передачи данных
	urlResult := make(chan words.Result)
	// Открываем файл с урлами
	urlFile, urlErr := os.Open("url.txt")
	if urlErr != nil {
		fmt.Println("Ошибка считывания файла с url:", urlErr)
	}
	// Открываем файл с результатами
	jsonFile, jsonErr := os.Create("result.json")
	if jsonErr != nil {
		fmt.Println("Ошибка считывания файла с url:", jsonErr)
	}

	defer urlFile.Close()
	defer jsonFile.Close()

	// Инициализируем сканер
	scanner := bufio.NewScanner(urlFile)
	// Проходимся по всем урлам в файле, для каждого определяем топ 3
	for scanner.Scan() {
		var tag string
		row := strings.Split(scanner.Text(), ",")
		url := row[0]
		// Если указан необходимый тег в файле
		if len(row) == 2 {
			tag = row[1]
		}
		go words.GetTopData(url, urlResult, tag)
		urlCount++
	}
	fmt.Println("Всего горутин ", urlCount)
	for ; urlCount != 0; urlCount-- {
		// Считываем результат из канала
		result := <-urlResult
		// Если нет результатов
		if result.Count == [3]int{0, 0, 0} {
			data = append(data, result)
		}
	}
	close(urlResult)
	// Запись результатов в файл
	encodeErr := result.PrettyWriteJSON(jsonFile, data)
	if encodeErr != nil {
		fmt.Println("Ошибка записи результата: ", encodeErr)
	}
	// Вывод затраченного времени
	fmt.Printf("Time spend: %v", time.Since(start))
}
