package top3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/fatih/camelcase"
)

// processWorker воркер для считывания урла и запуска top3.URL
func processWorker(resultChan chan Result, errChan chan error, urlChan chan string, downloadChan chan []byte, o ...Option) {
	for {
		// Получаем url из main горутины
		url, isOpen := <-urlChan
		if !isOpen {
			break
		}
		// Инициируем рейтинг
		rating := make(map[string]int)
		result := Result{
			Url: url,
		}
		for {
			// Получаем частями тело ответа
			bytes := <-downloadChan
			if bytes == nil {
				break
			}

			// Достаём из части текст
			text, err := extractText(bytes)
			if err != nil {
				err = fmt.Errorf("error: %v url: %v", err, url)
				errChan <- err
				return
			}

			// Дробим текст на слова
			words := spliter(text)

			// Проходимся по всем словам и считаем их количество
			for _, w := range words {
				var up = upCount(w)
				// Если заглавных букв больше 1 и это не абривиатура
				if up > 1 && up != utf8.RuneCountInString(w) {

					// ????? : Будут ли при этом слова проитерированы в этом цикле?

					words = append(words, camelcase.Split(w)...)
					continue
				}
				w = strings.ToLower(w)
				// Проверяем строку на количество символов и является ли она словом
				if utf8.RuneCountInString(w) > 3 && isWord(w) {
					// Считаем количество упомянаний
					rating[w] += 1
				} else {
					continue
				}

				changed := false

				for index, value := range result.Words {
					if value == w && value != "" {
						result.Count[index] += 1
						changed = true
						break
					}
				}

				if changed {
					continue
				}

				if rating[w] > result.Count[2] {
					result.Count[2] = rating[w]
					result.Words[2] = w
				}
				if result.Count[2] > result.Count[1] {
					result.Count[1], result.Count[2] = result.Count[2], result.Count[1]
					result.Words[1], result.Words[2] = result.Words[2], result.Words[1]
				}
				if result.Count[1] > result.Count[0] {
					result.Count[0], result.Count[1] = result.Count[1], result.Count[0]
					result.Words[0], result.Words[1] = result.Words[1], result.Words[0]
				}
			}

		}

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
	}
}

func downloadWorker(byteChan chan []byte, urlChan chan string, errChan chan error, client *http.Client) {

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
