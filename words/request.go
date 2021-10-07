package words

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
	Domains      *RepeatOptions
}

// getResponce отправляет запрос
func getResponce(u string, o ...SendReqOptions) (*http.Response, error) {
	var options SendReqOptions
	if len(o) > 0 {
		options = o[0]
	}
	var resp *http.Response

	un, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	domain := un.Hostname()
	DomainsMutex.Lock()
	_, domainContain := Domains[domain] // race
	DomainsMutex.Unlock()
	if domainContain {

		request, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return nil, err
		}

		request.Header = http.Header{
			"Authority":                 []string{domain},
			"Pragma":                    []string{"no-cache"},
			"Cache-control":             []string{"no-cache"},
			"Sec-ch-ua":                 []string{`"Google Chrome";v="89", "Chromium";v="89", ";Not A Brand";v="99"`},
			"Sec-ch-ua-mobile":          []string{"?0"},
			"Upgrade-insecure-requests": []string{"1"},
			"User-agent":                []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36"},
			"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			"Dnt":                       []string{"1"},
			"Sec-fetch-site":            []string{"none"},
			"Sec-fetch-mode":            []string{"navigate"},
			"Sec-fetch-user":            []string{"?1"},
			"Sec-fetch-dest":            []string{"document"},
		}
		DomainsMutex.Lock()
		domains := options.Domains.Domains[domain] // race
		DomainsMutex.Unlock()
		if domains > options.HostReqLimit {
			fmt.Println("Wait")
			for domains >= options.HostReqLimit {
				time.Sleep(100 * time.Millisecond)
				DomainsMutex.Lock()
				domains = options.Domains.Domains[domain]
				DomainsMutex.Unlock()
			}
		}

		DomainsMutex.Lock()
		options.Domains.Domains[domain] += 1 // race
		DomainsMutex.Unlock()

		resp, err := options.Client.Do(request)

		DomainsMutex.Lock()
		options.Domains.Domains[domain] -= 1
		DomainsMutex.Unlock()

		if err != nil {
			return resp, err
		}
		return resp, nil
	}
	resp, err = http.Get(u)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// sendRequest обрабатывает запрос
func sendRequest(url string, o ...SendReqOptions) (*http.Response, error) {

	var options SendReqOptions
	if len(o) > 0 {
		options = o[0]
	}

	resp, err := getResponce(url, options)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 503 || resp.StatusCode == 429 {

		fmt.Printf("'%v' was received. Attempt to resend the request %v \n", resp.Status, url)
		keepAlive := resp.Header.Values("Keep-Alive")
		var count int
		if len(keepAlive) > 0 {
			timeout := strings.ReplaceAll(keepAlive[0], "timeout=", "")
			count, err = strconv.Atoi(timeout)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// for name, values := range resp.Header {
			// 	for _, value := range values {
			// 		fmt.Println(name, value)
			// 	}
			// }
			count = 15
		}
		fmt.Printf("Timeout is %v seconds\n", count)
		num := 1
		for resp.StatusCode != 200 {
			time.Sleep(time.Duration(count) * time.Second)
			resp, err = getResponce(url, options)
			if err != nil {
				fmt.Println(err)
			}
			if resp.StatusCode == 404 {
				break
			}
			fmt.Printf("Try %v for %v status %v \n", num, url, resp.Status)
			num++
		}
	}

	if resp.StatusCode != 200 {
		// keepAlive := resp.Header.Values("keep-alive")
		return nil, fmt.Errorf("http: %v ", resp.Status)
	}
	return resp, nil
}
