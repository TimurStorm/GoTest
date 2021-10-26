package top3

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
)

type TestResponce struct {
	Url        string
	Domain     string
	StatusCode int
	Options    SendReqOptions
}

var TestResponceData = []TestResponce{
	{
		Url:        "https://habr.com/ru/post/325364/",
		Domain:     "habr.com",
		StatusCode: 200,
	},
	{
		Url:        "https://fabulae.ru/prose_b.php?id=117958",
		Domain:     "fabulae.ru",
		StatusCode: 200,
		Options: SendReqOptions{
			HostReqLimit: 1,
			Client:       new(http.Client),
		},
	},
	{
		Url:        "https://fabulae.ru/prose_b.php?id=117958",
		Domain:     "fabulae.ru",
		StatusCode: 200,
		Options: SendReqOptions{
			HostReqLimit: 1,
			Client:       new(http.Client),
		},
	},
	{
		Url:        "http://dgsdgsgseg.com/",
		Domain:     "dgsdgsgseg.com",
		StatusCode: 404,
	},
}

func TestSendRequest(t *testing.T) {
	HostsMutex.Lock()
	Hosts["fabulae.ru"] = 0
	HostsMutex.Unlock()
	var wg sync.WaitGroup
	for _, test := range TestResponceData {
		wg.Add(1)
		go func(test TestResponce) {
			defer wg.Done()
			resp, err := sendRequest(test.Url, test.Domain, test.Options)
			if err != nil {
				fmt.Println(err)
				return
			}
			if resp.StatusCode != test.StatusCode {
				t.Errorf("Статус коды не совпадают! Для сайта %v был получен статус %v", test.Url, resp.StatusCode)
			}
		}(test)
	}
	wg.Wait()
}

func TestGetResponceBody(t *testing.T) {
	HostsMutex.Lock()
	Hosts["www.avito.ru"] = 0
	HostsMutex.Unlock()

	file, err := os.Open("url_test.txt")
	if err != nil {
		fmt.Printf("Файл с урлами не найден")
	}
	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		err := scanner.Err()
		if err != nil {
			fmt.Println("Ошибка чтения файла")
			break
		}
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			resp, err := getResponceBody(url, SendReqOptions{Client: new(http.Client), HostReqLimit: 5})
			if err != nil {
				t.Errorf("Ошибка отправки запроса: %v", err)
				return
			}
			if resp.StatusCode != 200 {
				fmt.Println(resp.Status)
			}
		}(url)
	}
	wg.Wait()
}
