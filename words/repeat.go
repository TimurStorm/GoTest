package words

import (
	"bufio"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"
)

// getRepeatedHosts возвращает все хосты из файла urlFileName с упоминаем более 1
func getRepeatedHosts(urlFileName string) ([]string, error) {
	// Результат
	var result []string

	repeatedDomains := make(map[string]int)
	// Открываем файл с урлами
	urlFile, err := os.Open(urlFileName)
	if err != nil {
		return nil, err
	}
	defer urlFile.Close()

	// Считываем урлы и считаем количество повторений хостов
	scanner := bufio.NewScanner(urlFile)
	for scanner.Scan() {
		u := scanner.Text()
		err := scanner.Err()
		if err != nil {
			return nil, err
		}
		if u != "" {
			un, err := url.Parse(u)
			if err != nil {
				return nil, err
			}
			domain := un.Hostname()
			repeatedDomains[domain] += 1
		}
	}
	// Отсеиваем все уникальные хосты
	for domain, count := range repeatedDomains {
		if count > 1 {
			result = append(result, domain)
		}
	}
	return result, nil
}

// getUnique воздвращает все уникальны слова массива строк
func getUnique(allWords []string) []string {
	// Уникальные слова
	var result []string

	// Отбираем все слова с длинной больше 3
	for _, word := range allWords {
		if utf8.RuneCountInString(word) > 3 && isWord(word) && !arrayContainString(result, word) {
			result = append(result, word)
		}
	}
	return result
}

// getPopularWords возращает классификацию слов по популярности в виде map и наибольшее количество поторений
func getPopularWords(text string) (map[int][]string, int) {

	// Результат
	result := make(map[int][]string)

	// Получаем все слова из текста
	allWords := spliter(text, " -.,?!()<>_")

	// Наибольшее количество упомянаний
	var maxCount int = 0

	// Получаем уникальные слова
	uniqueWords := getUnique(allWords)

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
