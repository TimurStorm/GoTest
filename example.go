package main

import (
	"fmt"
	"main/top3"
	"runtime"
	"time"
)

func exampleForFile() {
	start := time.Now()
	err := top3.ForFile("url.txt", "result.json")
	if err != nil {
		panic(err)
	}
	since := time.Since(start)
	fmt.Println("Без воркеров", since)
}

func exampleForFileWithWorkers() {
	start := time.Now()
	err := top3.ForFile("url.txt", "result.json", top3.WithWriteWorkers(3), top3.WithProcessWorkers(2))
	if err != nil {
		panic(err)
	}
	since := time.Since(start)
	fmt.Println("С воркерами", since)
}

func main() {
	// Используем 1 ядро процессора
	runtime.GOMAXPROCS(1)

	exampleForFile()
	exampleForFileWithWorkers()
	// Вывод затраченного времени

}
