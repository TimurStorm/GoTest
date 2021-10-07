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
	err := words.GetTopForFile("url.txt", "result.json", words.GetTopFFOptions{Tags: []string{"p", "a"}, HostReqLimit: 10})

	if err != nil {
		fmt.Println(err)
	}
	result, _ := words.GetTop("https://www.habr.com")
	fmt.Println(result)

	// Вывод затраченного времени
	fmt.Printf("Time spend: %v", time.Since(start))
}
