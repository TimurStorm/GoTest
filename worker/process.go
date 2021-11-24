package worker

import (
	"fmt"
	"main/support"
	"strings"
	"unicode/utf8"

	"github.com/fatih/camelcase"
)

// processWorker воркер для считывания урла и запуска top3.URL
func Process(resultChan chan Result, errChan chan error, urlChan chan string, downloadChan chan []byte, tags ...string) {
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
			text, err := support.ExtractText(bytes)
			if err != nil {
				err = fmt.Errorf("error: %v url: %v", err, url)
				errChan <- err
				return
			}

			// Дробим текст на слова
			words := support.Spliter(text)

			// Проходимся по всем словам и считаем их количество
			for _, w := range words {
				var up = support.UpCount(w)
				// Если заглавных букв больше 1 и это не абривиатура
				if up > 1 && up != utf8.RuneCountInString(w) {

					// ????? : Будут ли при этом слова проитерированы в этом цикле?

					words = append(words, camelcase.Split(w)...)
					continue
				}
				w = strings.ToLower(w)
				// Проверяем строку на количество символов и является ли она словом
				if utf8.RuneCountInString(w) > 3 && support.IsWord(w) {
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
