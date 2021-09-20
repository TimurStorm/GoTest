package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"main/result"
	"main/words"
)

func main() {
	// Группа ожидания для горутин
	var wg sync.WaitGroup
	// Используем 1 ядро процессора
	runtime.GOMAXPROCS(1)
	// Для измерения времени
	start := time.Now()
	// Массив результирующих данных
	var data []words.Result
	// Открываем файл с урлами
	urlFile, urlErr := os.Open("stres_test.txt")
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
		wg.Add(1)
		var tag string
		row := strings.Split(scanner.Text(), ",")
		url := row[0]
		// Если указан необходимый тег в файле
		if len(row) == 2 {
			tag = row[1]
		}
		go words.GetTopData(url, &data, wg, tag)
	}

	// Ждём выолнения всех горутин
	wg.Wait()
	// Запись результатов в файл
	encodeErr := result.PrettyWriteJSON(jsonFile, data)
	if encodeErr != nil {
		fmt.Println("Ошибка записи результата: ", encodeErr)
	}
	// Вывод затраченного времени
	fmt.Printf("Time spend: %v", time.Since(start))
}
