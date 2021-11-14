# GoTest


## Описание:
Библеотека для получения топ 3 популярных слова на указанном сайте.

## Использование:
### Топ 3 с одного сайта:
```top3.URL("example.com")```

top3.URL возвращает 3 самых популярных слова на странице и количество их упоминаний.

### Опции:
```WithClient(c http.Client)``` - для кастомного клиента
```WithTags(t []string)``` - для получения информации из указанных тегов
```WithAttemptCount(c uint)``` - макимальное количество попыток отправить повторно запрос после получения ошибки
```WithAttemptTimeout(t time.Duration)``` - время между попытками

### Топ 3 из текста:
```top3.ForText("Некоторый текст")```

top3.ForText выстраивает рейтинг популярных слов и возвращает 3 самых популярных слова на странице с количеством упоминаний. 

### Топ 3 из файла с урлами:
```top3.ForFile("urls.txt", "responce.json")```

top3.ForFile cчитывает урлы из urls.txt и для каждого запускает top3.URL. Успешный результат записывается в виде строки в файл responce.json в формате Simple Json.

### Опции:
```WithClient(c http.Client)``` - для кастомного клиента
```WithTags(t []string)``` - для получения информации из указанных тегов
```WithHostReqLimit(lim uint)``` - ограничение количества запросов на один хост
```WithAttemptCount(c uint)``` - макимальное количество попыток отправить повторно запрос после получения ошибки
```WithAttemptTimeout(t time.Duration)``` - время между попытками
```WithWriteWorkers(w uint)``` - количество воркеров на запись результатов в файл
```WithProcessWorkers(w uint)```- количество воркеров на обработку урлов из файла

## Полезные команды:
### Старт:
* `go run example.go`                                      - запуск примера

### Тесты:
* `go test -coverprofile cov.out ./top3 `               - покрытие тестами
* `go tool cover -html cov.out`                         - просмотр покрытия в бразуере
* `go tool cover -func cov.out`                         - просмотр покрытия в консоли

### Линтеры:
* `gocritic check main.go` - go-critic
* `golint main.go`         - golint
* `go vet main.go`         - vet
* `go fmt main.go`         - fmt

## Зависимости:
1) [goquery](github.com/PuerkitoBio/goquery) 
2) [html2text](jaytaylor.com/html2text)
3) [gocritic](https://github.com/go-critic/go-critic) 

