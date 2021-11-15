package top3

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func limitDo(domain string, request *http.Request, options *AllOptions) (*http.Response, error) {

	// Для облегчения читаймости
	hosts := options.hosts

	// Проверяем наличие хоста в мапе повторяющихся
	// Получаем количество подключений к данному хосту
	hostCount, contain := hosts.Map[domain]

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
		hosts.MapMutex.RLock()
		hosts.Map[domain] += 1
		hosts.MapMutex.RUnlock()

		// Отправка запроса
		resp, err := options.Client.Do(request)

		// - 1 запрос
		hosts.MapMutex.RLock()
		hosts.Map[domain] -= 1
		hosts.MapMutex.RUnlock()

		if err != nil {
			return resp, err
		}
		return resp, nil
	}
}

// sendRequest отправляет запрос
func sendRequest(u string, domain string, options *AllOptions) (*http.Response, error) {

	var resp *http.Response

	// Создаём запрос
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	// Если нет ограничений на отправку запросов на 1 хост
	// Просто отправляем запрос
	if options.HostReqLimit == 0 {

		resp, err = options.Client.Do(request)
		if err != nil {
			return resp, err
		}
		//
	} else {
		resp, err = limitDo(domain, request, options)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func getTimeout(resp *http.Response, defaultTimeout time.Duration) (time.Duration, error) {
	// Получаем timeout
	retryAfter := resp.Header.Values("Retry-After")

	// Timeout
	var timeout = defaultTimeout
	var err error
	if timeout == 0 {
		timeout = 7 * time.Second
	}

	// Если был найден Retry-After в заголовках ответа
	if len(retryAfter) > 0 {
		count, err := strconv.Atoi(retryAfter[0])
		if err == nil {
			timeout = time.Duration(count) * time.Second
			return timeout, nil
		}

		t, err := time.Parse(time.RFC1123, retryAfter[0])
		if err == nil {
			timeout = time.Since(t)
			return timeout, nil
		}
	}
	return timeout, err
}

// tryResend совершает попытку повторной отправки запроса
func tryResend(resp *http.Response, u string, host string, options *AllOptions) (*http.Response, error) {
	var (
		newResp *http.Response
		err     error
		attempt uint
	)

	fmt.Printf("'%v' было получено. Попытка повторного отправки запроса на %v \n", resp.Status, u)

	// Инициализируем timeout
	// По умолчанию равен 7, может быть изменён из опцией AttemptTimeout
	timeout, err := getTimeout(resp, options.AttemptTimeout)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Таймаут - %v\n", timeout)

	// Пытаемся получить хороший ответ от сервера
	for resp.StatusCode != 200 {
		if options.AttemptCount != 0 && options.AttemptCount < attempt {
			break
		}
		time.Sleep(timeout)
		newResp, err = sendRequest(u, host, options)
		if err != nil {
			return newResp, err
		}

		// Если была не найдена старница, невалиден запрос, запрещён доступ к ресурсу
		if resp.StatusCode == 404 || resp.StatusCode == 400 || resp.StatusCode == 403 {
			break
		}
		fmt.Printf("Попытка %v для %v статус %v \n", attempt, u, resp.Status)
		attempt++
	}
	return newResp, nil
}

// getResponceBody обрабатывает запрос
func getResponceBody(u string, o ...Option) ([]byte, error) {

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
	resp, err := sendRequest(u, host, options)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// В случае если было отправлено большое количество запросов в ближайшее время
	if resp.StatusCode == 503 || resp.StatusCode == 429 {

		resp, err = tryResend(resp, u, host, options)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("http: %v ", resp.Status)
		}
	}

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respByte, nil
}
