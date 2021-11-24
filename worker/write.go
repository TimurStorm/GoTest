package worker

import (
	"encoding/json"
	"fmt"
)

type Result struct {
	Url   string
	Words [3]string
	Count [3]int
}

// writeWorker воркер для записи данных
func Write(resultChan chan Result, errChan chan error, encoder *json.Encoder) {
	for {
		// Ждём результат
		result, isOpen := <-resultChan
		if !isOpen {
			break
		}

		// Производим запись
		err := encoder.Encode(result)
		if err != nil {
			err = fmt.Errorf("error: %v url: %v", err, result.Url)
			fmt.Println(err)
			errChan <- err
			return
		}
	}
}
