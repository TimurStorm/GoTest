package support

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"jaytaylor.com/html2text"
)

// extractText возвращает текст запроса
func ExtractText(responceData []byte, tags ...string) (string, error) {
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
