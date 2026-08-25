package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func decodeJSON(
	request *http.Request,
	target interface{},
) error {
	if !isJSON(request) {
		return errors.New("请求需要 application/json")
	}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("请求体只能包含一个 JSON 对象")
	}
	return nil
}
func queryTime(
	request *http.Request,
	key string,
	fallback time.Time,
) time.Time {
	value := strings.TrimSpace(request.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.Parse(
		time.RFC3339, value)
	if err != nil {
		parsed, err = time.Parse("2006-01-02", value)
	}
	if err != nil {
		return fallback
	}
	return parsed.UTC()
}
func queryInt(
	request *http.Request,
	key string,
	fallback int,
) int {
	value := request.URL.Query().Get(key)
	var parsed int
	if _, err := fmtSscanf(value, &parsed); err != nil {
		return fallback
	}
	return parsed
}
func fmtSscanf(
	value string,
	target *int,
) (int, error) {
	return fmt.Sscanf(value, "%d", target)
}

func queryFloat(
	request *http.Request,
	key string,
	fallback float64,
) float64 {
	value := strings.TrimSpace(request.URL.Query().Get(key))
	parsed, err :=
		strconv.ParseFloat(
			value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func queryBool(
	request *http.Request,
	key string,
) bool {
	value, err :=
		strconv.ParseBool(
			strings.TrimSpace(request.URL.Query().Get(key)))
	return err == nil && value
}
