package main

import (
	"fmt"
	"main/top3"
	"runtime"
	"time"
)

func exampleTopFile() {
	err := top3.GetTopFile("url.txt", "result.json", top3.WithTags([]string{"p", "a"}))
	if err != nil {
		panic(err)
	}
}

func exampleTopURL() {
	result, err := top3.GetTop("https://habr.com/ru/feed/")
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func main() {
	// Используем 1 ядро процессора
	runtime.GOMAXPROCS(1)
	// Для измерения времени
	start := time.Now()
	exampleTopFile()
	exampleTopURL()
	// Вывод затраченного времени
	fmt.Printf("Time spend: %v", time.Since(start))
}
