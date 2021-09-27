package words

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
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

type Options struct {
	Tags       []string
	SiteRepeat bool
}

// GetTopWords Возвращает топ 3 слов текста по упоминаниям и их количество
func GetTopWords(text string) ([3]string, [3]int, error) {
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

	return resWords, resCount, nil
}

// GetText возвращает текст запроса
func GetText(responce *http.Response, params ...[]string) (string, error) {
	// Результат
	var result string

	// Ошибка извлечения тектса из тега
	var divErr error

	// Html теги
	var tags []string
	if len(params[0]) > 0 {
		tags = params[0]
	} else {
		tags = append(tags, "div")
	}

	// Вспомогательная функция:
	textFrom := func(html string) error {
		text, err := html2text.FromString(html, html2text.Options{OmitLinks: true})
		result += text
		return err
	}

	// Считываем тело запроса
	doc, docErr := goquery.NewDocumentFromReader(responce.Body)
	if docErr != nil {
		return "", docErr
	}
	// Получаем разметку тега body
	bodyHTML, err := doc.Html()
	// Ошибка конвертации html
	if err != nil {
		return "", err
	}
	if strings.Contains(bodyHTML, "div") {
		// Для каждого тега в файле получаем его html-вёрстку, из которой получаем текст
		for _, tag := range tags {
			doc.Find(tag).Each(func(index int, item *goquery.Selection) {
				html, err := item.Html()
				if err != nil {
					divErr = nil
				}
				if !(strings.Contains(html, "div")) {
					err = textFrom(html)
					if err != nil {
						divErr = nil
					}
				}
			})
		}
	} else {
		textFrom(bodyHTML)
	}

	if divErr != nil {
		return "", divErr
	}
	return result, nil
}

// GetTop возвращает результат с топ-3 наиболее упоминаемых слов и их колиечеством на странице сайта
func GetTop(url string, o ...Options) (Result, error) {
	var options Options
	if len(o) > 0 {
		options = o[0]
	}
	fmt.Printf("REQUEST %v \n", url)

	// Отправляем запрос
	resp, err := http.Get(url)
	if err != nil {
		return Result{}, err
	}

	if resp.StatusCode != 200 {
		return Result{}, errors.New("Bad responce status: " + resp.Status)
	}

	// Инициализируем теги
	var tags []string
	if len(options.Tags) > 0 {
		tags = options.Tags
	}

	// Получаем текст
	text, err := GetText(resp, tags)
	if err != nil {
		return Result{}, err
	}

	// Получаем топ 3 слова
	words, count, err := GetTopWords(text)
	if err != nil {
		return Result{}, err
	}

	result := Result{Url: url, Count: count, Words: words}

	return result, nil
}

// FindTopForFile сканирует файл urlFileName и для каждого url производит GetTop. Результат записывается в resultFileName
func FindTopForFile(urlFileName string, resultFileName string, o ...Options) error {
	var options Options
	if len(o) > 0 {
		options = o[0]
	}

	// Открываем файл с урлами
	urlFile, err := os.Open(urlFileName)
	if err != nil {
		return err
	}
	defer urlFile.Close()

	// Открываем файл с результатами
	resultFile, err := os.Create(resultFileName)
	if err != nil {
		return err
	}

	defer resultFile.Close()
	defer urlFile.Close()

	encoder := json.NewEncoder(resultFile)

	// Инициализируем сканер
	scanner := bufio.NewScanner(urlFile)

	var wg errgroup.Group
	// Проходимся по всем урлам в файле, для каждого определяем топ 3
	for scanner.Scan() {
		url := scanner.Text()
		err := scanner.Err()
		if err != nil {
			return err
		}
		if url != "" {
			wg.Go(func() error {
				// Получаем результат
				var result Result
				result, err = GetTop(url, options)

				if err != nil {
					return err
				}

				// Записываем результат
				err = encoder.Encode(result)
				if err != nil {
					return err
				}
				return nil
			})
		}
	}
	err = wg.Wait()
	if err != nil {
		return err
	}

	return nil
}
