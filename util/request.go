package util

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
)

func FetchData(url string, v any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Warn("fail to get data", "resp", resp)
		return fmt.Errorf("fail to get data")
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bodyBytes, v); err != nil {
		return err
	}

	return nil
}
