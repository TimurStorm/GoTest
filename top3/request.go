package top3

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SendReqOptions struct {
	Client       *http.Client
	HostReqLimit int
}

// sendRequest отправляет запрос
func sendRequest(u string, domain string, o ...SendReqOptions) (*http.Response, error) {
	var options SendReqOptions
	if len(o) > 0 {
		options = o[0]
	}
	var resp *http.Response

	// Проверяем наличие хоста в мапе повторяющихся
	HostsMutex.Lock()
	_, domainContain := Hosts[domain]
	HostsMutex.Unlock()

	// Если хост в мапе повторяющихся
	if domainContain && options.HostReqLimit != 0 {

		// Получаем количество подключений к данному хосту
		HostsMutex.Lock()
		hostCount := Hosts[domain]
		HostsMutex.Unlock()

		// Ожидание пока не освободится место новому запросу
		for hostCount >= options.HostReqLimit {
			time.Sleep(100 * time.Millisecond)
			HostsMutex.Lock()
			hostCount = Hosts[domain]
			HostsMutex.Unlock()
		}

		// + 1 запрос
		HostsMutex.Lock()
		Hosts[domain] += 1
		HostsMutex.Unlock()

		// Отправка запроса
		resp, err := http.Get(u)

		// - 1 запрос
		HostsMutex.Lock()
		Hosts[domain] -= 1
		HostsMutex.Unlock()

		if err != nil {
			return resp, err
		}
		return resp, nil
	}
	// Стандартный метод отправки
	resp, err := http.Get(u)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// getResponceBody обрабатывает запрос
func getResponceBody(u string, o ...SendReqOptions) (*http.Response, error) {
	var options SendReqOptions
	if len(o) > 0 {
		options = o[0]
	}

	// Получаем хост
	un, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	// Хост сайта
	host := un.Hostname()

	// Получаем ответ
	resp, err := sendRequest(u, host, options)
	if err != nil {
		return nil, err
	}

	// В случае если было отправлено большое количество запросов в ближайшее время
	if resp.StatusCode == 503 || resp.StatusCode == 429 {
		fmt.Printf("'%v' was received. Attempt to resend the request %v \n", resp.Status, u)

		// Получаем timeout
		keepAlive := resp.Header.Values("Keep-Alive")
		retryAfter := resp.Header.Values("Retry-After")

		// Timeout
		var count int

		if len(keepAlive) > 0 {
			// Если был найден Keep-Alive в заголовках ответа

			timeout := strings.ReplaceAll(keepAlive[0], "timeout=", "")
			count, err = strconv.Atoi(timeout)
			if err != nil {
				fmt.Println(err)
			}

		} else if len(retryAfter) > 0 {
			// Если был найден Retry-After в заголовках ответа

			count, err = strconv.Atoi(retryAfter[0])
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// По умолчанию
			count = 60
		}
		fmt.Printf("Timeout is %v seconds\n", count)

		// Счётчик попыток
		num := 1

		// Пытаемся получить хороший ответ от сервера
		for resp.StatusCode != 200 {
			time.Sleep(time.Duration(count) * time.Second)
			resp, err = sendRequest(u, host, options)
			if err != nil {
				fmt.Println(err)
			}

			// Если была не найдена старница, невалиден запрос, запрещён доступ к ресурсу
			if resp.StatusCode == 404 || resp.StatusCode == 400 || resp.StatusCode == 403 {
				break
			}
			fmt.Printf("Try %v for %v status %v \n", num, u, resp.Status)
			num++
		}
	}

	if resp.StatusCode != 200 {
		// keepAlive := resp.Header.Values("keep-alive")
		return nil, fmt.Errorf("http: %v ", resp.Status)
	}
	return resp, nil
}
