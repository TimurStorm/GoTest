package main

import (
	"fmt"
	"main/words"
	"runtime"
	"time"
)

func main() {
	// Используем 1 ядро процессора
	runtime.GOMAXPROCS(1)

	// Для измерения времени
	start := time.Now()
	err := words.FindTopForFile("url.txt", "result.json", words.Options{Tags: []string{"p", "a"}, SiteRepeat: true})

	if err != nil {
		fmt.Println(err)
	}

	// Вывод затраченного времени
	fmt.Printf("Time spend: %v", time.Since(start))
}
