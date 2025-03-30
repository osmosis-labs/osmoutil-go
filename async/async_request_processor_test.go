package async_test

import (
	"context"
	"testing"
	"time"

	"github.com/osmosis-labs/osmoutil-go/async"
	"github.com/osmosis-labs/osmoutil-go/retry"
	"github.com/stretchr/testify/require"
)

// Define simple test types to use in our tests
type TestInput struct {
	Value string
}

type TestOutput struct {
	ProcessedValue string
	ProcessedAt    time.Time
}

const defaultMaxDuration = 10 * time.Second

var (
	defaultRetryConfig = &retry.RetryConfig{
		MaxDuration: 30 * time.Second,
	}
)

func TestAsyncRequestProcessor(t *testing.T) {
	// Common test setup function to reduce duplication
	setupRequestProcessor := func(processor async.RequestProcessor[TestInput, TestOutput]) *async.AsyncRequestProcessor[TestInput, TestOutput] {
		return async.NewAsyncRequstProcessor[TestInput, TestOutput](10, processor, defaultRetryConfig, defaultMaxDuration)
	}

	// Helper to create a test request
	createRequest := func(id string, value string) async.Request[TestInput] {
		return async.Request[TestInput]{
			ID:        id,
			CreatedAt: time.Now(),
			Data:      TestInput{Value: value},
		}
	}

	// Basic processor that just prepends "processed:" to the input
	basicProcessor := async.FunctionProcessor[TestInput, TestOutput]{
		ProcessFn: func(ctx context.Context, req async.Request[TestInput]) (TestOutput, error) {
			return TestOutput{
				ProcessedValue: "processed:" + req.Data.Value,
				ProcessedAt:    time.Now(),
			}, nil
		},
	}

	// Test cases
	tests := []struct {
		name           string
		processor      async.RequestProcessor[TestInput, TestOutput]
		requests       []async.Request[TestInput]
		expectedOutput func(req async.Request[TestInput]) string
	}{
		{
			name:      "Single request processing",
			processor: basicProcessor,
			requests: []async.Request[TestInput]{
				createRequest("req-1", "test-data"),
			},
			expectedOutput: func(req async.Request[TestInput]) string {
				return "processed:" + req.Data.Value
			},
		},
		{
			name:      "Multiple requests processing",
			processor: basicProcessor,
			requests: []async.Request[TestInput]{
				createRequest("req-1", "data-1"),
				createRequest("req-2", "data-2"),
				createRequest("req-3", "data-3"),
			},
			expectedOutput: func(req async.Request[TestInput]) string {
				return "processed:" + req.Data.Value
			},
		},
		{
			name: "Custom processor logic",
			processor: async.FunctionProcessor[TestInput, TestOutput]{
				ProcessFn: func(ctx context.Context, req async.Request[TestInput]) (TestOutput, error) {
					return TestOutput{
						ProcessedValue: req.Data.Value + "-transformed",
						ProcessedAt:    time.Now(),
					}, nil
				},
			},
			requests: []async.Request[TestInput]{
				createRequest("req-special", "custom"),
			},
			expectedOutput: func(req async.Request[TestInput]) string {
				return req.Data.Value + "-transformed"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and start worker
			worker := setupRequestProcessor(tt.processor)
			worker.Start()
			defer worker.Stop() // Ensure the worker is stopped after the test

			// Create a map to track expected responses
			expected := make(map[string]string)
			for _, req := range tt.requests {
				expected[req.ID] = tt.expectedOutput(req)
			}

			// Create a channel to collect responses
			responsesChan := make(chan async.Response[TestOutput], len(tt.requests))

			// Start a goroutine to collect responses
			go func() {
				for resp := range worker.Responses() {
					responsesChan <- resp
					// If we've collected all expected responses, we can break
					if len(responsesChan) == len(tt.requests) {
						break
					}
				}
			}()

			// Submit all requests
			for _, req := range tt.requests {
				success := worker.Submit(req)
				require.True(t, success, "Failed to submit request: %s", req.ID)
			}

			// Collect and verify all responses
			responses := make(map[string]async.Response[TestOutput])

			// Wait for all expected responses with a timeout
			timeout := time.After(3 * time.Second)
			for i := 0; i < len(tt.requests); i++ {
				select {
				case resp := <-responsesChan:
					responses[resp.RequestID] = resp
				case <-timeout:
					t.Fatalf("Test timed out waiting for responses")
				}
			}

			// Verify we got responses for all requests
			require.Equal(t, len(tt.requests), len(responses), "Did not receive expected number of responses")

			// Verify each response
			for id, expectedOutput := range expected {
				resp, exists := responses[id]
				require.True(t, exists, "No response received for request ID: %s", id)
				require.NoError(t, resp.Error, "Response contained an error: %v", resp.Error)
				require.Equal(t, expectedOutput, resp.Data.ProcessedValue, "Response data did not match expected for request ID: %s", id)

				// Verify response has a reasonable duration
				require.True(t, resp.Duration > 0, "Response duration should be greater than 0")
				require.True(t, resp.Duration < 3*time.Second, "Response duration was too long: %v", resp.Duration)

				// Verify ProcessedAt time is reasonable (between start of test and now)
				testStart := time.Now().Add(-3 * time.Second)
				require.True(t, resp.Data.ProcessedAt.After(testStart),
					"ProcessedAt time is too old: %v", resp.Data.ProcessedAt)
				require.True(t, resp.Data.ProcessedAt.Before(time.Now().Add(1*time.Second)),
					"ProcessedAt time is in the future: %v", resp.Data.ProcessedAt)
			}
		})
	}
}

func TestAsyncRequestProcessorWithFunc(t *testing.T) {
	// Test that the NewWorkerWithFunc constructor works correctly

	processFn := func(ctx context.Context, req async.Request[string]) (int, error) {
		return len(req.Data), nil
	}

	worker := async.NewAsyncRequestWorkerWithFunc[string, int](5, defaultMaxDuration, defaultRetryConfig, processFn)

	// Start the worker
	worker.Start()
	defer worker.Stop()

	// Submit a request
	req := async.Request[string]{
		ID:        "func-test",
		CreatedAt: time.Now(),
		Data:      "hello world",
	}

	success := worker.Submit(req)
	require.True(t, success, "Failed to submit request")

	// Collect the response
	var resp async.Response[int]
	select {
	case resp = <-worker.Responses():
		// Got a response
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for response")
	}

	// Verify the response
	require.Equal(t, "func-test", resp.RequestID, "Response ID did not match request ID")
	require.NoError(t, resp.Error, "Response contained an error")
	require.Equal(t, 11, resp.Data, "Response data did not match expected length")
}

func TestWorkerChannelFull(t *testing.T) {
	// Create a worker with very small buffer and a processor that takes time
	processor := async.FunctionProcessor[string, string]{
		ProcessFn: func(ctx context.Context, req async.Request[string]) (string, error) {
			time.Sleep(50 * time.Millisecond) // Slow processor
			return req.Data, nil
		},
	}

	// Buffer size of 2
	worker := async.NewAsyncRequstProcessor(2, processor, defaultRetryConfig, defaultMaxDuration)
	worker.Start()
	defer worker.Stop()

	// Submit requests quickly to fill the buffer
	results := make([]bool, 5)

	for i := 0; i < 5; i++ {
		req := async.Request[string]{
			ID:        "req-buffer-" + string(rune('0'+i)),
			CreatedAt: time.Now(),
			Data:      "test",
		}

		results[i] = worker.Submit(req)

		// Don't wait between submissions to try to fill the buffer
	}

	// At least some submissions should fail when buffer is full
	atLeastOneFailed := false
	for _, success := range results {
		if !success {
			atLeastOneFailed = true
			break
		}
	}

	require.True(t, atLeastOneFailed, "Expected at least one submission to fail when buffer is full")
}

func TestWorkerStopProcessingRemainingItems(t *testing.T) {
	// Test that when Stop is called, remaining items in the queue are processed

	processedItems := make(map[string]bool)

	processor := async.FunctionProcessor[string, string]{
		ProcessFn: func(ctx context.Context, req async.Request[string]) (string, error) {
			processedItems[req.ID] = true
			return req.Data, nil
		},
	}

	worker := async.NewAsyncRequstProcessor[string, string](10, processor, defaultRetryConfig, defaultMaxDuration)
	worker.Start()

	// Submit several requests
	requests := []string{"stop-1", "stop-2", "stop-3", "stop-4", "stop-5"}

	for _, id := range requests {
		req := async.Request[string]{
			ID:        id,
			CreatedAt: time.Now(),
			Data:      "test",
		}

		worker.Submit(req)
	}

	// Immediately stop the worker
	worker.Stop()

	// Verify that all requests were processed
	for _, id := range requests {
		require.True(t, processedItems[id], "Request was not processed: %s", id)
	}
}
