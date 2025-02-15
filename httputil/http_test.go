package httputil_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/osmosis-labs/osmoutil-go/httputil"
	"github.com/stretchr/testify/require"
)

type TestResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func TestMakeRequest(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test request headers
		require.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

		// For POST requests, verify Content-Type
		if r.Method == http.MethodPost {
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		}

		// Return test response
		resp := TestResponse{
			Message: "success",
			Status:  "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	ctx := context.Background()

	// Test POST request with payload
	t.Run("POST with payload", func(t *testing.T) {
		payload := map[string]string{"test": "data"}
		headers := map[string]string{"X-Test-Header": "test-value"}
		var response TestResponse

		_, err := httputil.Post(
			ctx,
			server.URL,
			payload,
			headers,
			&response,
		)
		require.NoError(t, err)
		require.Equal(t, "success", response.Message)
		require.Equal(t, "ok", response.Status)
	})

	// Test GET request without payload
	t.Run("GET without payload", func(t *testing.T) {
		headers := map[string]string{"X-Test-Header": "test-value"}
		var response TestResponse

		_, err := httputil.Get(
			ctx,
			server.URL,
			headers,
			&response,
		)
		require.NoError(t, err)
		require.Equal(t, "success", response.Message)
		require.Equal(t, "ok", response.Status)
	})

	// Test error cases
	t.Run("invalid URL", func(t *testing.T) {
		_, err := httputil.Get(
			ctx,
			"invalid-url",
			nil,
			nil,
		)
		require.Error(t, err)
	})

	t.Run("use lowercase headers", func(t *testing.T) {
		headers := map[string]string{"x-test-header": "test-value"}
		var response TestResponse

		_, err := httputil.Get(ctx, server.URL, headers, &response)
		require.NoError(t, err)
		require.Equal(t, "success", response.Message)
		require.Equal(t, "ok", response.Status)
	})
}

func TestBuildURLWithParams(t *testing.T) {
	tests := []struct {
		name      string
		urlPrefix string
		endpoint  string
		params    map[string]string
		want      string
		wantErr   bool
	}{
		{
			name:      "basic URL without params",
			urlPrefix: "https://api.example.com",
			endpoint:  "/v1/data",
			params:    nil,
			want:      "https://api.example.com/v1/data",
			wantErr:   false,
		},
		{
			name:      "URL with single param",
			urlPrefix: "https://api.example.com",
			endpoint:  "/v1/data",
			params:    map[string]string{"key": "value"},
			want:      "https://api.example.com/v1/data?key=value",
			wantErr:   false,
		},
		{
			name:      "URL with multiple params",
			urlPrefix: "https://api.example.com",
			endpoint:  "/v1/data",
			params:    map[string]string{"key1": "value1", "key2": "value2"},
			want:      "https://api.example.com/v1/data?key1=value1&key2=value2",
			wantErr:   false,
		},
		{
			name:      "invalid URL",
			urlPrefix: "://invalid",
			endpoint:  "/test",
			params:    nil,
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httputil.BuildURLWithParams(tt.urlPrefix, tt.endpoint, tt.params)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
