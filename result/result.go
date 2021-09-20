package result

import (
	"encoding/json"
	"io"
	"main/words"
)

// PrettyWriteJSON записывает в .json файл результирующие данные
func PrettyWriteJSON(jsonFile io.Writer, data []words.Result) error {
	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "    ")
	encodeErr := encoder.Encode(data)
	if encodeErr != nil {
		return encodeErr
	}
	return nil
}
