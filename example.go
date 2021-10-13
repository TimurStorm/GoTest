package main

import (
	"fmt"
	"runtime"
	"time"

	"main/top3"
)

func main() {
	// Используем 1 ядро процессора
	runtime.GOMAXPROCS(1)

	// Для измерения времени
	start := time.Now()
	err := top3.GetTopForFile("url.txt", "result.json", top3.GetTopOptions{Tags: []string{"p", "a"}, HostReqLimit: 10})
	if err != nil {
		fmt.Println(err)
	}
	result, err := top3.GetTop("https://habr.com/ru/post/578414/")

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)

	// Вывод затраченного времени
	fmt.Printf("Time spend: %v", time.Since(start))
}