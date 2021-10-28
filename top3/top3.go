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

type AllOptions struct {
	Tags         []string
	HostReqLimit int
	Client       http.Client
}

type Option func(*AllOptions)

func WithTags(t []string) Option {
	return func(opts *AllOptions) {
		opts.Tags = t
	}
}

func WithHostReqLimit(lim int) Option {
	return func(opts *AllOptions) {
		opts.HostReqLimit = lim
	}
}

func WithClient(c http.Client) Option {
	return func(opts *AllOptions) {
		opts.Client = c
	}
}

// GetPopularWords возвращает топ 3 слов текста по упоминаниям и их количество
func GetPopularWords(text string) ([3]string, [3]int, error) {

	// Итоговые топ 3 слова
	var resWords [3]string

	// Итоговое число упомянаний топ 3 слов
	var resCount [3]int

	// Определяем популярные слова и максимальное их значение
	wordCount, maxCount := getRating(text)

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

// extractText возвращает текст запроса
func extractText(responce *http.Response, tags ...string) (string, error) {
	// Результат
	var result string
	// Теги
	var t []string
	if len(tags) > 0 {
		t = tags
	}

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
	if strings.Contains(bodyHTML, "div") && len(t) > 0 {
		// Для каждого тега в файле получаем его html-вёрстку, из которой получаем текст
		for _, tag := range t {
			doc.Find(tag).Each(func(index int, item *goquery.Selection) {
				html, err := item.Html()
				if err != nil {
					divErr = nil
				}
				if !(strings.Contains(html, "div")) && !(strings.Contains(html, "script")) {
					err = textFrom(html)
					if err != nil {
						divErr = nil
					}
				}
			})
		}
		if divErr != nil {
			return "", divErr
		}
	} else {
		textFrom(bodyHTML)
	}

	return result, nil
}

// GetTop возвращает результат с топ-3 наиболее упоминаемых слов и их количеством на странице сайта
func GetTop(url string, o ...Option) (Result, error) {
	options := &AllOptions{}
	for _, opt := range o {
		opt(options)
	}
	fmt.Printf("REQUEST %v \n", url)

	// Отправляем запрос
	resp, err := getResponceBody(url, WithClient(options.Client), WithHostReqLimit(options.HostReqLimit))
	if err != nil {
		return Result{}, err
	}

	// Получаем текст из запроса
	text, err := extractText(resp, options.Tags...)
	if err != nil {
		return Result{}, err
	}

	// Получаем топ 3 слова
	words, count, err := GetPopularWords(text)
	if err != nil {
		return Result{}, err
	}

	result := Result{Url: url, Count: count, Words: words}

	return result, nil
}

// GetTopFile сканирует файл urlFileName и для каждого url производит GetTop. Результат записывается в resultFileName
func GetTopFile(urlFileName string, resultFileName string, o ...Option) error {
	options := &AllOptions{}
	for _, opt := range o {
		opt(options)
	}

	if options.Client.Timeout == 0 {
		options.Client.Timeout = 5 * time.Second
	}

	fmt.Println(options.Client)

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
	// Инициализируем енкодер
	encoder := json.NewEncoder(resultFile)

	// Инициализируем сканер
	scanner := bufio.NewScanner(urlFile)

	// Инициализируем errgroup
	var wg errgroup.Group

	// Проходимся по всем урлам в файле, для каждого определяем топ 3
	for scanner.Scan() {
		url := scanner.Text()

		// Если была обнаружена ошибка при считывании
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
					result, err = GetTop(url, o...)
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
