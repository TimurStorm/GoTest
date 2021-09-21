package words

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"jaytaylor.com/html2text"
)

type Result struct {
	Url   string
	Words [3]string
	Count [3]int
}

func spliter(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	splitter := func(r rune) bool {
		return m[r] == 1
	}

	return strings.FieldsFunc(s, splitter)
}

// GetTopData записывает в массив интерфейс с топ 3 словами и их количеством на странице
func GetTopData(url string, data *[]Result, wg *sync.WaitGroup, defaultTag ...string) {
	var tag string

	if defaultTag[0] != "" {
		tag = defaultTag[0]
	}

	fmt.Printf("REQUEST %v \n", url)
	// Отправляем запрос
	resp, respErr := http.Get(url)

	if respErr != nil {
		fmt.Println("Ошибка отправки запроса: ", respErr)
	} else if resp.StatusCode == 200 {
		text, textErr := getText(resp, tag)
		if textErr != nil {
			fmt.Println("Ошибка получения текста: ", textErr)
		} else {
			// Получаем топ 3 упомянаемых слова с колиством упомянаний
			words, count := getWordsCount(text)
			// Сохраняем полученные данные
			result := Result{url, words, count}
			*data = append(*data, result)

		}
	}
	wg.Done()
}

// getWordsCount Возвращает топ 3 слов текста
func getWordsCount(text string) ([3]string, [3]int) {
	// Вспомогательная функция: проверка на принадлежность строки к массиву
	containString := func(list []string, substing string) bool {
		for _, value := range list {
			if value == substing {
				return true
			}
		}
		return false
	}

	// Вспомогательная функция: проверка на слово
	isWord := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}

	// Уникальные слова
	var data []string

	// Наибольшее количество упомянаний
	var maxCount int = 0

	// Итоговые топ 3 слова
	var resWords [3]string

	// Итоговое число упомянаний топ 3 слов
	var resCount [3]int

	// Для определения самых популярных
	wordCount := make(map[int][]string)

	// Получаем все слова из текста
	allWords := spliter(text, " -.,?!()<>_")

	// Определяем все уникальные слова
	for _, word := range allWords {
		if utf8.RuneCountInString(word) > 3 && isWord(word) && !containString(data, word) {
			data = append(data, word)
		}
	}

	// Классификация слов по популярности
	for _, word := range data {
		c := strings.Count(text, word)
		wordCount[c] = append(wordCount[c], word)
		if c > maxCount {
			maxCount = c
		}
	}

	// Отбираем топ 3 слова
	for nodeCount := 0; nodeCount < 3 && maxCount != 0; {
		if _, ok := wordCount[maxCount]; ok {
			for _, word := range wordCount[maxCount] {
				if nodeCount == 3 {
					break
				}
				resWords[nodeCount] = word
				resCount[nodeCount] = maxCount
				nodeCount++
			}
		}
		maxCount--
	}

	return resWords, resCount
}

// getText возвращает текст запроса
func getText(responce *http.Response, defaultTag ...string) (string, error) {

	// Результат
	var result string

	// Ошибка получения текста
	var textErr, htmlErr error

	// Тег для поиска информации
	var tag string

	textFrom := func(html string) {
		text, err := html2text.FromString(html, html2text.Options{OmitLinks: true})
		textErr = err
		result += text
	}

	// Если был получен тег по умолчанию
	if defaultTag[0] != "" {
		tag = strings.ReplaceAll(defaultTag[0], " ", "")
	}

	// Считываем тело запроса
	doc, docErr := goquery.NewDocumentFromReader(responce.Body)
	if docErr != nil {
		return "", docErr
	}
	// Получаем разметку тега body
	bodyHTML, htmlErr := doc.Html()
	// Ошибка конвертации html
	if htmlErr != nil {
		return "", htmlErr
	}
	if strings.Contains(bodyHTML, "div") {
		// Считываем тег
		if tag == "" {
			fmt.Printf("Введите тег для получения конкретной информации с (например: 'a') или 'body' для полной информации:")
			fmt.Scan(&tag)
		}

		// Для каждого тега в файле получаем его html-вёрстку, из которой получаем текст
		doc.Find(tag).Each(func(index int, item *goquery.Selection) {
			html, err := item.Html()
			htmlErr = err
			if tag == "div" {
				if !(strings.Contains(html, "div")) {
					textFrom(html)
				}
			} else {
				textFrom(html)
			}

		})
	} else {
		textFrom(bodyHTML)
	}

	// Ошибка конвертации html
	if htmlErr != nil {
		return "", htmlErr
	}

	// Ошибка передачи текста
	if textErr != nil {
		return "", textErr
	}

	return result, nil
}
