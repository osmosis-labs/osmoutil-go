package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type httpMethod string

const (
	HttpGET  httpMethod = http.MethodGet
	HttpPOST httpMethod = http.MethodPost
)

// makeRequest handles common HTTP request functionality by creating and executing an HTTP request
// with the provided method, URL, and optional payload. If response is provided, the response body
// will be JSON decoded into it.
func makeRequest(ctx context.Context, method httpMethod, url string, payload interface{}, headers map[string]string, response interface{}) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
		body = bytes.NewReader(jsonPayload)
	}

	req, err := http.NewRequestWithContext(ctx, string(method), url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// If response interface is provided, decode JSON directly into it
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return nil, nil
	}

	// Otherwise, return the raw response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, nil
}

// BuildURLWithParams creates a URL with query parameters by combining a base URL prefix,
// endpoint path, and optional query parameters.
func BuildURLWithParams(urlPrefix, endpoint string, params map[string]string) (string, error) {
	baseURL, err := url.Parse(urlPrefix + endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	if len(params) > 0 {
		query := url.Values{}
		for key, value := range params {
			query.Add(key, value)
		}
		baseURL.RawQuery = query.Encode()
	}

	return baseURL.String(), nil
}

// Get is a convenience wrapper for making HTTP GET requests
func Get(ctx context.Context, url string, headers map[string]string, response interface{}) ([]byte, error) {
	return makeRequest(ctx, HttpGET, url, nil, headers, response)
}

// Post is a convenience wrapper for making HTTP POST requests
func Post(ctx context.Context, url string, payload interface{}, headers map[string]string, response interface{}) ([]byte, error) {
	return makeRequest(ctx, HttpPOST, url, payload, headers, response)
}
