package main

import (
	"fmt"
	"main/top3"
	"runtime"
	"time"
)

func main() {
	// Используем 1 ядро процессора
	runtime.GOMAXPROCS(1)
	// Для измерения времени
	start := time.Now()
	err := top3.GetTopFile("url.txt", "result.json", top3.WithTags([]string{"p", "a"}))
	if err != nil {
		fmt.Println(err)
	}

	// Вывод затраченного времени
	fmt.Printf("Time spend: %v", time.Since(start))
}
