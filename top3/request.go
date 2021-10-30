package top3

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/manucorporat/try"
)

// sendRequest отправляет запрос
func sendRequest(u string, domain string, o ...Option) (*http.Response, error) {

	options := &AllOptions{}
	for _, opt := range o {
		opt(options)
	}

	// Создаём запрос
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	//Если нет ограничений на отправку запросов на 1 хост
	//Просто отправляем запрос
	if options.HostReqLimit == 0 {

		resp, err := options.Client.Do(request)
		if err != nil {
			return resp, err
		}
		return resp, nil

	} else {

		//Для облегчения читаймости
		hosts := options.hosts

		// Проверяем наличие хоста в мапе повторяющихся
		// Получаем количество подключений к данному хосту
		hosts.MapMutex.Lock()
		hostCount, contain := hosts.Map[domain]
		hosts.MapMutex.Unlock()

		// Если не в мапе
		if !contain {
			resp, err := options.Client.Do(request)
			if err != nil {
				return resp, err
			}

			return resp, nil
		} else {

			// Ожидание пока не освободится место новому запросу
			for hostCount >= options.HostReqLimit {
				time.Sleep(100 * time.Millisecond)
				hosts.MapMutex.Lock()
				hostCount = hosts.Map[domain]
				hosts.MapMutex.Unlock()
			}

			// + 1 запрос
			hosts.MapMutex.Lock()
			hosts.Map[domain] += 1
			hosts.MapMutex.Unlock()

			// Отправка запроса
			resp, err := options.Client.Do(request)

			// - 1 запрос
			hosts.MapMutex.Lock()
			hosts.Map[domain] -= 1
			hosts.MapMutex.Unlock()

			if err != nil {
				return resp, err
			}
			return resp, nil
		}
	}
}

// getResponceBody обрабатывает запрос
func getResponceBody(u string, o ...Option) (*http.Response, error) {

	options := &AllOptions{}
	for _, opt := range o {
		opt(options)
	}

	// Получаем хост
	un, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	// Хост сайта
	host := un.Hostname()

	// Получаем ответ
	resp, err := sendRequest(u, host, o...)
	if err != nil {
		return nil, err
	}

	// В случае если было отправлено большое количество запросов в ближайшее время
	if resp.StatusCode == 503 || resp.StatusCode == 429 {
		fmt.Printf("'%v' было получено. Попытка повторного отправки запроса на %v \n", resp.Status, u)

		// Получаем timeout
		retryAfter := resp.Header.Values("Retry-After")

		// Timeout
		var timeout = options.AttemptTimeout
		if timeout == 0 {
			timeout = 15 * time.Second
		}

		// Если был найден Retry-After в заголовках ответа
		if len(retryAfter) > 0 {
			try.This(func() {
				count, err := strconv.Atoi(retryAfter[0])
				if err != nil {
					panic(err)
				}
				timeout = time.Duration(count) * time.Second
				fmt.Println(timeout)

			}).Catch(func(e try.E) {
				t, err := time.Parse(time.RFC1123, retryAfter[0])
				if err != nil {
					panic(err)
				}
				timeout = time.Since(t)
				fmt.Println(timeout)
			})
		}
		fmt.Printf("Таймаут - %v\n", timeout)

		// Счётчик попыток
		var num uint = 0

		// Пытаемся получить хороший ответ от сервера
		for resp.StatusCode != 200 {
			if options.AttemptCount != 0 && options.AttemptCount < num {
				break
			}
			time.Sleep(timeout)
			resp, err = sendRequest(u, host, o...)
			if err != nil {
				return nil, err
			}

			// Если была не найдена старница, невалиден запрос, запрещён доступ к ресурсу
			if resp.StatusCode == 404 || resp.StatusCode == 400 || resp.StatusCode == 403 {
				break
			}
			fmt.Printf("Попытка %v для %v статус %v \n", num, u, resp.Status)
			num++
		}

		if resp.StatusCode != 200 {
			// keepAlive := resp.Header.Values("keep-alive")
			return nil, fmt.Errorf("http: %v ", resp.Status)
		}
	}
	return resp, nil
}
