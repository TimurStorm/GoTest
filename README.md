# GoTest
### Useful commands:
```golang
go run main.go                               - start programm
go test -coverprofile cov.out ./words        - test coverage
go tool cover -html cov.out                  - view coverage in browser
go tool cover -func cov.out                  - view coverage in console
```
### Dependencies:
1) [goquery](github.com/PuerkitoBio/goquery) 
2) [html2text](jaytaylor.com/html2text)
3) [gocritic](https://github.com/go-critic/go-critic) 

### Linters:
* `gocritic check main.go` - go-critic
* `golint main.go`         - golint
* `go vet main.go`         - vet
* `go fmt main.go`         - fmt
