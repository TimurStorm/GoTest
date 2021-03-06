package top3

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"jaytaylor.com/html2text"
)

type hosts struct {
	MapMutex *sync.RWMutex
	Map      map[string]uint
}

type Result struct {
	Url   string
	Words [3]string
	Count [3]int
}

type AllOptions struct {
	Tags           []string
	Client         http.Client
	hosts          *hosts
	WriteWorkers   uint
	ProcessWorkers uint
	HostReqLimit   uint
	AttemptCount   uint
	AttemptTimeout time.Duration
}

type Option func(*AllOptions)

func WithTags(t []string) Option {
	return func(opts *AllOptions) {
		opts.Tags = t
	}
}

func WithHostReqLimit(lim uint) Option {
	return func(opts *AllOptions) {
		opts.HostReqLimit = lim
	}
}

func WithClient(c http.Client) Option {
	return func(opts *AllOptions) {
		opts.Client = c
	}
}

func withHosts(h *hosts) Option {
	return func(opts *AllOptions) {
		opts.hosts = h
	}
}

func WithAttemptCount(c uint) Option {
	return func(opts *AllOptions) {
		opts.AttemptCount = c
	}
}

func WithAttemptTimeout(t time.Duration) Option {
	return func(opts *AllOptions) {
		opts.AttemptTimeout = t
	}
}

func WithWriteWorkers(w uint) Option {
	return func(opts *AllOptions) {
		opts.WriteWorkers = w
	}
}

func WithProcessWorkers(w uint) Option {
	return func(opts *AllOptions) {
		opts.ProcessWorkers = w
	}
}

// ForText возвращает топ 3 слов текста по упоминаниям и их количество
func ForText(text string) ([3]string, [3]int, error) {

	// Результат
	var (
		words [3]string
		count [3]int
	)

	// Определяем популярные слова и максимальное их значение
	rating, max := getRating(text)

	// Отбираем топ 3 слова
	// node - количество записей ( не больше 3)
	for node := 0; node < 3 && max != 0; {
		if _, ok := rating[max]; ok {
			for _, word := range rating[max] {
				if node == 3 {
					break
				}
				words[node] = word
				count[node] = max
				node++
			}
		}
		max--
	}

	return words, count, nil
}

// extractText возвращает текст запроса
func extractText(responceData []byte, tags ...string) (string, error) {
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
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(responceData))
	if err != nil {
		return "", err
	}

	// Получаем разметку тега body
	bodyHTML, err := doc.Html()
	if err != nil {
		return "", err
	}
	if strings.Contains(bodyHTML, "div") && len(tags) > 0 {
		tagFilter := func(html string) {
			if !(strings.Contains(html, "div")) && !(strings.Contains(html, "script")) {
				err = textFrom(html)
				if err != nil {
					divErr = nil
				}
			}
		}
		// Для каждого тега в файле получаем его html-вёрстку, из которой получаем текст
		for _, tag := range tags {
			doc.Find(tag).Each(func(index int, item *goquery.Selection) {
				html, err := item.Html()
				if err != nil {
					divErr = nil
				}
				tagFilter(html)
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

// ForPage возвращает результат с топ-3 наиболее упоминаемых слов и их количеством на странице сайта
func ForPage(url string, o ...Option) (Result, error) {
	options := &AllOptions{}
	for _, opt := range o {
		opt(options)
	}
	fmt.Printf("REQUEST %v \n", url)
	// Отправляем запрос
	resp, err := getResponceBody(url, WithClient(options.Client), WithHostReqLimit(options.HostReqLimit), withHosts(options.hosts))
	if err != nil {
		return Result{Url: url}, err
	}

	// Получаем текст из запроса
	text, err := extractText(resp, options.Tags...)
	if err != nil {
		return Result{Url: url}, err
	}

	// Получаем топ 3 слова
	words, count, err := ForText(text)
	if err != nil {
		return Result{Url: url}, err
	}

	result := Result{Url: url, Count: count, Words: words}

	return result, nil
}

// ForFile сканирует файл urlFileName и для каждого url производит URL. Результат записывается в resultFileName
func ForFile(urlFileName string, resultFileName string, o ...Option) error {
	// Считываем опции
	options := &AllOptions{
		Client:       http.Client{Timeout: 5 * time.Second},
		WriteWorkers: 1,
	}
	for _, opt := range o {
		opt(options)
	}

	// Таймаут по умолчанию
	// TODO: сделать более точную настройку клиента

	// Если задан лимит запросов на хост
	if options.HostReqLimit != 0 {
		options.hosts = &hosts{MapMutex: new(sync.RWMutex), Map: make(map[string]uint)}
		repeated, err := getRepeatedHosts(urlFileName)
		if err != nil {
			return err
		}
		for _, host := range repeated {
			options.hosts.Map[host] = 0
		}
	}
	o = append(o, withHosts(options.hosts))

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

	// Инициализируем енкодер
	encoder := json.NewEncoder(resultFile)

	// Считаем количество строк в файле
	rowCount, err := lineCounter(urlFileName)
	if err != nil {
		return err
	}
	rowCount *= 2

	// Инициализируем каналы для ошибок и урлов
	errChan := make(chan error, rowCount)
	resultChan := make(chan Result, rowCount)

	defer close(resultChan)
	defer close(errChan)

	// Запускаем воркеры для записи, по умолчанию 1
	var w uint

	for w < options.WriteWorkers {
		// Воркер для записи
		go writeWorker(resultChan, errChan, encoder)
		w += 1
	}

	// Стандартный режим
	if options.ProcessWorkers == 0 {
		scanner := bufio.NewScanner(urlFile)
		for scanner.Scan() {
			url := scanner.Text()
			err := scanner.Err()
			if err != nil {
				return err
			}
			go process(url, resultChan, errChan, o...)
		}
		// Режим с воркерами
	} else {
		// Инициализируем ридер
		reader := bufio.NewReader(urlFile)
		// Запускаем воркеры
		var w uint
		for w < options.ProcessWorkers {
			// Воркер для считывания и обработки
			go processWorker(resultChan, errChan, reader, o...)
			w += 1
		}
	}

	// Ждём выполнения всех процессов и воркеров
	for {
		err = <-errChan
		if err != nil {
			return err
		}
		rowCount -= 1
		if rowCount == 0 {
			break
		}
	}

	return nil
}

// process обёртка для top3.URL с каналами resultChan и errChan
func process(url string, resultChan chan Result, errChan chan error, o ...Option) {

	result, err := ForPage(url, o...)
	if err != nil {
		errChan <- err
		return
	}

	resultChan <- result
	errChan <- nil
}
