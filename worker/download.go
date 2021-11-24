package worker

import (
	"bufio"
	"io"
	"net/http"
)

func Download(byteChan chan []byte, urlChan chan string, errChan chan error, client *http.Client) {

	for {
		// Получаем урл
		url, isOpen := <-urlChan
		if !isOpen {
			break
		}

		// Формируем запрос
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			errChan <- err
			return
		}

		// Отпарвляем запрос
		res, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer res.Body.Close()

		// Инициализируем ридер
		reader := bufio.NewReader(res.Body)
		for {
			// Считываем построчно тело запроса
			bytes, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				errChan <- err
				return
			}
			// Отправляем байты в processWorker
			byteChan <- bytes
			if err == io.EOF {
				byteChan <- nil
				break
			}

		}
	}
}
