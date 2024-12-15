package httputil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RunGet is a helper function to make an HTTP GET request and decode the response into a struct.
func RunGet(ctx context.Context, url string, headers map[string]string, response interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}
