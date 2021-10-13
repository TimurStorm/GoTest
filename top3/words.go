package top3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"jaytaylor.com/html2text"
)

var HostsMutex = new(sync.Mutex)
var Hosts = make(map[string]int)

type Result struct {
	Url   string
	Words [3]string
	Count [3]int
}

type GetTopOptions struct {
	Tags         []string
	HostReqLimit int
	Client       http.Client
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
func GetText(responce *http.Response, tags ...[]string) (string, error) {
	// Результат
	var result string

	// Ошибка извлечения тектса из тега
	var divErr error

	// Вспомогательная функция:
	textFrom := func(html string) error {
		text, err := html2text.FromString(html, html2text.Options{OmitLinks: true})
		result += text
		return err
	}

	// Считываем тело запроса
	doc, err := goquery.NewDocumentFromReader(responce.Body)
	if err != nil {
		return "", err
	}

	// Получаем разметку тега body
	bodyHTML, err := doc.Html()
	if err != nil {
		return "", err
	}

	if strings.Contains(bodyHTML, "div") {
		// Теги
		var t []string
		if len(tags[0]) > 0 {
			t = tags[0]
		} else {
			t = append(t, "div")
		}
		// Для каждого тега в файле получаем его html-вёрстку, из которой получаем текст
		for _, tag := range t {
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
func GetTop(url string, o ...GetTopOptions) (Result, error) {
	var options GetTopOptions
	if len(o) > 0 {
		options = o[0]
	}
	fmt.Printf("REQUEST %v \n", url)

	// Отправляем запрос
	resp, err := sendRequest(url, SendReqOptions{Client: &options.Client, HostReqLimit: options.HostReqLimit})
	if err != nil {
		return Result{}, err
	}

	// Получаем текст
	text, err := GetText(resp, options.Tags)
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

// GetTopForFile сканирует файл urlFileName и для каждого url производит GetTop. Результат записывается в resultFileName
func GetTopForFile(urlFileName string, resultFileName string, o ...GetTopOptions) error {
	var options GetTopOptions
	if len(o) > 0 {
		options = o[0]
	}
	// Устанавливаем клиент
	options.Client = http.Client{Timeout: time.Duration(5) * time.Second}

	// Если задан лимит запросов на один хост
	if options.HostReqLimit != 0 {
		repeated, err := getRepeatedHosts(urlFileName)
		if err != nil {
			return err
		}
		for _, host := range repeated {
			Hosts[host] = 0
		}
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

	// Инициализируем errgroup
	var wg errgroup.Group

	// Проходимся по всем урлам в файле, для каждого определяем топ 3
	for scanner.Scan() {
		url := scanner.Text()
		err := scanner.Err()
		if err != nil {
			return err
		}
		if url != "" {
			wg.Go(
				func() error {
					// Получаем результат
					var result Result
					var err error
					result, err = GetTop(url, options)
					if err != nil {
						err = fmt.Errorf("error: %v url: %v", err, url)
						fmt.Println(err)
						return err
					}

					// Записываем результат
					err = encoder.Encode(result)
					if err != nil {
						err = fmt.Errorf("error: %v url: %v", err, url)
						fmt.Println(err)
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