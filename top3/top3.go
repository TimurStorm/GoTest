package top3

import (
	"bufio"
	"encoding/json"
	"main/worker"
	"net/http"
	"os"
	"sync"
	"time"
)

type hosts struct {
	MapMutex *sync.RWMutex
	Map      map[string]uint
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

// ForFile сканирует файл urlFileName и для каждого url производит URL. Результат записывается в resultFileName
func ForFile(urlFileName string, resultFileName string, o ...Option) error {
	// Считываем опции
	options := &AllOptions{
		Client:         http.Client{Timeout: 5 * time.Second},
		WriteWorkers:   1,
		ProcessWorkers: 1,
	}
	for _, opt := range o {
		opt(options)
	}

	// ВРЕМЕННО: исключено из приложения
	// Если задан лимит запросов на хост
	// if options.HostReqLimit != 0 {
	// 	options.hosts = &hosts{MapMutex: new(sync.RWMutex), Map: make(map[string]uint)}
	// 	repeated, err := support.GetRepeatedHosts(urlFileName)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	for _, host := range repeated {
	// 		options.hosts.Map[host] = 0
	// 	}
	// }
	//o = append(o, withHosts(options.hosts))

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
	// rowCount, err := lineCounter(urlFileName)
	// if err != nil {
	// 	return err
	// }
	// rowCount *= 2

	// Инициализируем каналы для ошибок и урлов
	errChan := make(chan error)
	urlProcessChan := make(chan string)
	urlDownloadChan := make(chan string)
	wgDone := make(chan bool)
	resultChan := make(chan worker.Result)

	// Запускаем воркеры для записи, по умолчанию 1
	go worker.Write(resultChan, errChan, encoder)

	// Инициализируем ридер
	scanner := bufio.NewScanner(urlFile)
	var w uint
	var wg sync.WaitGroup

	// Запуск воркеров-процессов
	for w = 0; w < options.ProcessWorkers; w++ {
		wg.Add(2)
		bytesDownloadChan := make(chan []byte)
		go func() {
			worker.Process(resultChan, errChan, urlProcessChan, bytesDownloadChan, options.Tags...)
			wg.Done()
		}()
		go func() {
			worker.Download(bytesDownloadChan, urlDownloadChan, errChan, &options.Client)
			wg.Done()
		}()
	}
	for scanner.Scan() {
		url := scanner.Text()
		err = scanner.Err()
		if err != nil {
			return err
		}
		urlDownloadChan <- url
		urlProcessChan <- url
	}
	close(urlDownloadChan)
	close(urlProcessChan)

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-errChan:
		close(errChan)
		return err
	}

	return nil
}
