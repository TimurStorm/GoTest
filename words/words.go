package words

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"jaytaylor.com/html2text"
)

func GetWordsCount(text string) ([3]string, [3]int) {
	// Вспомогательная функция: проверка на принадлежность строки к массиву
	contain_string := func(list []string, substing string) bool {
		for _, value := range list {
			if value == substing {
				return true
			}
		}
		return false
	}

	// Вспомогательная функция: проверка на слово
	is_word := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}

	// Вспомогательная функция: делители для разделения текста на слова
	split := func(r rune) bool {
		return r == ':' || r == '.' || r == ' ' || r == ',' || r == '!' || r == '?' || r == '-' || r == '<' || r == '>' || r == '_' || r == '(' || r == ')'
	}

	//Для уникальных слов
	var data []string

	//Наибольшее количество упомянаний
	var max_count int = 0

	//Итоговые топ 3 слова
	var res_words [3]string

	//Итоговое число упомянаний топ 3 слов
	var res_count [3]int

	//Для определения самых популярных
	word_count := make(map[int][]string)

	//Получаем все слова из текста
	all_words := strings.FieldsFunc(text, split)

	//Определяем все уникальные слова
	for _, word := range all_words {
		if utf8.RuneCountInString(word) > 3 && is_word(word) && !contain_string(data, word) {
			data = append(data, word)
		}
	}

	//Классификация слов по популярности
	for _, word := range data {
		c := strings.Count(text, word)
		word_count[c] = append(word_count[c], word)
		if c > max_count {
			max_count = c
		}
	}

	//Отбираем топ 3 слова
	for node_count := 0; node_count < 3 && max_count != 0; {
		if _, ok := word_count[max_count]; ok {
			for _, word := range word_count[max_count] {
				if node_count == 3 {
					break
				}
				res_words[node_count] = word
				res_count[node_count] = max_count
				node_count++
			}
		}
		max_count--
	}

	return res_words, res_count
}

func GetText(responce *http.Response, default_tags string) string {
	//Результат
	var text string

	//Тег для поиска информации
	var tag string = default_tags

	//Считываем тело запроса
	doc, _ := goquery.NewDocumentFromReader(responce.Body)

	//Считываем тег
	fmt.Println("Введите тег для получения конкретной информации (например: 'a') или 'body' для полной информации:")
	fmt.Scan(&tag)

	//Для каждого тега в файле получаем его html-вёрстку, из которой
	doc.Find(tag).Each(func(index int, item *goquery.Selection) {
		html, _ := item.Html()
		t, _ := html2text.FromString(html, html2text.Options{OmitLinks: true})
		text += t
	})
	return text
}
