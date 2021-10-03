package words

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"jaytaylor.com/html2text"
)

type Result struct {
	Url   string
	Words [3]string
	Count [3]int
}

type Options struct {
	Tags       []string
	SiteRepeat bool
}

type OptionsPlus struct {
	Tags        []string
	Client      http.Client
	MaxTryCount int
}

// GetTopWords Возвращает топ 3 слов текста по упоминаниям и их количество
func GetTopWords(text string) ([3]string, [3]int, error) {

	// Итоговые топ 3 слова
	var resWords [3]string

	// Итоговое число упомянаний топ 3 слов
	var resCount [3]int

	// Определяем популярные слова и максимальное их значение
	wordCount, maxCount := getPopularWords(text)

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

// GetTop возвращает результат с топ-3 наиболее упоминаемых слов и их количеством на странице сайта
func GetTop(url string, o ...OptionsPlus) (Result, error) {
	var options OptionsPlus
	if len(o) > 0 {
		options = o[0]
	}
	fmt.Printf("REQUEST %v \n", url)
	// Отправляем запрос
	resp, err := sendRequest(url, &options.Client)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

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

	client := http.Client{}
	var repeatedDomains []string
	if options.SiteRepeat {
		repeated, err := getRepeated(urlFileName)
		if err != nil {
			return err
		}
		repeatedDomains = repeated
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
			domain, err := getDomain(url)
			if err != nil {
				fmt.Println(err)
			}
			wg.Go(
				func() error {
					// Получаем результат
					var result Result
					if arrayContainString(repeatedDomains, domain) {
						result, err = GetTop(url, OptionsPlus{Tags: options.Tags, Client: client})
					} else {
						result, err = GetTop(url, OptionsPlus{Tags: options.Tags})
					}
					if err != nil {
						fmt.Println(err)
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
