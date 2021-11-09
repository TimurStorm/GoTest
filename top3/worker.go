package top3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// processWorker воркер для считывания урла и запуска top3.URL
func processWorker(resultChan chan Result, errChan chan error, reader *bufio.Reader, o ...Option) {
	for {
		// Считываем урл
		url, err := readln(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			err = fmt.Errorf("error: %v url: %v", err, url)
			errChan <- err
			return
		}

		result, err := URL(url, o...)

		if err != nil {
			err = fmt.Errorf("error: %v url: %v", err, url)
			errChan <- err
			return
		}

		// Отправляем результат main горутине, что всё хорошо
		errChan <- nil
		// Отправляем результат writeWorker горутине
		resultChan <- result
	}
}

// writeWorker воркер для записи данных
func writeWorker(resultChan chan Result, errChan chan error, encoder *json.Encoder) {
	for {
		// Ждём результат
		result, isOpen := <-resultChan
		if !isOpen {
			break
		}

		// Производим запись
		err := encoder.Encode(result)
		if err != nil {
			err = fmt.Errorf("error: %v url: %v", err, result.Url)
			fmt.Println(err)
			errChan <- err
			return
		}
		// Отправляем результат main горутине, что всё хорошо
		errChan <- nil
	}
}
