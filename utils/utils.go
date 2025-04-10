package utils

import "encoding/json"

func FancyJson[T any](data []byte) (T, error) {
	var result T

	err := json.Unmarshal(data, &result)

	return result, err
}
