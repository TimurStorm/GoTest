package words

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func spliter(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	splitter := func(r rune) bool {
		return m[r] == 1
	}

	return strings.FieldsFunc(s, splitter)
}

func getResponce(url string, client *http.Client) (*http.Response, error) {
	var resp *http.Response
	if client != http.DefaultClient {
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		domain, err := getDomain(url)
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
			"Sccept-language":           []string{"en-GB,en;q=0.9"},
		}
		resp, err := client.Do(request)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func sendRequest(url string, client *http.Client) (*http.Response, error) {

	resp, err := getResponce(url, client)
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
			resp, err = getResponce(url, client)
			if err != nil {
				fmt.Println(err)
			}
			if resp.StatusCode == 404 {
				break
			}
			fmt.Printf("Try %v for %v STATUS %v \n", num, url, resp.Status)
			num++
		}
	}

	if resp.StatusCode != 200 {
		//keepAlive := resp.Header.Values("keep-alive")
		return nil, errors.New(url + " " + resp.Status)
	}
	return resp, nil
}

// isWord проверка на слово
func isWord(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func getDomain(url string) (string, error) {
	words := spliter(url, "/:")
	if len(words) < 2 {
		return "", errors.New("NoDomain")
	}
	return words[1], nil
}

func getUnique(allWords []string) []string {
	// Уникальные слова
	var result []string
	for _, word := range allWords {
		if utf8.RuneCountInString(word) > 3 && isWord(word) && !arrayContainString(result, word) {
			result = append(result, word)
		}
	}
	return result
}

// arrayContainString проверяет на принадлежность строки к массиву
func arrayContainString(list []string, substing string) bool {
	for _, value := range list {
		if value == substing {
			return true
		}
	}
	return false
}

func getPopularWords(text string) (map[int][]string, int) {
	// Получаем все слова из текста
	allWords := spliter(text, " -.,?!()<>_")
	// Наибольшее количество упомянаний
	var maxCount int = 0
	uniqueWords := getUnique(allWords)
	result := make(map[int][]string)
	// Классификация слов по популярности
	for _, word := range uniqueWords {
		c := strings.Count(text, word)
		result[c] = append(result[c], word)
		if c > maxCount {
			maxCount = c
		}
	}
	return result, maxCount
}

func getRepeated(urlFileName string) ([]string, error) {
	// Результат
	var result []string

	uniquerepeatedDomains := make(map[string]int)
	// Открываем файл с урлами
	urlFile, err := os.Open(urlFileName)
	if err != nil {
		return nil, err
	}
	defer urlFile.Close()

	scanner := bufio.NewScanner(urlFile)
	for scanner.Scan() {
		url := scanner.Text()
		err := scanner.Err()
		if err != nil {
			return nil, err
		}
		if url != "" {
			domain, err := getDomain(url)
			if err != nil {
				fmt.Println(err)
			}
			uniquerepeatedDomains[domain] += 1
		}
	}
	for domain, count := range uniquerepeatedDomains {
		if count > 1 {
			result = append(result, domain)
		}
	}
	return result, nil
}
