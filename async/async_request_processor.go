package async

import (
	"context"
	"sync"
	"time"

	"github.com/osmosis-labs/osmoutil-go/retry"
)

// Request represents a work item to be processed by the worker
type Request[T any] struct {
	ID        string
	Data      T
	CreatedAt time.Time
}

// Response represents the outcome of processing a request
type Response[R any] struct {
	RequestID string
	Data      R
	Error     error
	Duration  time.Duration
}

// RequestProcessor defines the interface for custom request processors
type RequestProcessor[T any, R any] interface {
	Process(ctx context.Context, req Request[T]) (R, error)
}

// FunctionProcessor is an adapter that allows using functions as RequestProcessors
type FunctionProcessor[T any, R any] struct {
	ProcessFn func(ctx context.Context, req Request[T]) (R, error)
}

// Process implements the RequestProcessor interface
func (f FunctionProcessor[T, R]) Process(ctx context.Context, req Request[T]) (R, error) {
	return f.ProcessFn(ctx, req)
}

// AsyncRequestProcessor handles the processing of requests in a synchronous manner.
// Clients can submit requests to the processor and receive responses asynchronously.
type AsyncRequestProcessor[T any, R any] struct {
	requestChan  chan Request[T]
	responseChan chan Response[R]
	processor    RequestProcessor[T, R]
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	maxRetries   int

	retryConfig *retry.RetryConfig
}

// NewAsyncRequstProcessor creates a new background worker with the specified buffer size and processor
// If retryConfig is nil, no retry logic will be used
func NewAsyncRequstProcessor[T any, R any](
	bufferSize int,
	processor RequestProcessor[T, R],
	retryConfig *retry.RetryConfig,
) *AsyncRequestProcessor[T, R] {
	ctx, cancel := context.WithCancel(context.Background())

	return &AsyncRequestProcessor[T, R]{
		requestChan:  make(chan Request[T], bufferSize),
		responseChan: make(chan Response[R], bufferSize),
		processor:    processor,
		ctx:          ctx,
		cancel:       cancel,
		retryConfig:  retryConfig,
	}
}

var (
	// NoRetryConfig is a retry config that will not retry any requests
	NoRetryConfig *retry.RetryConfig = nil
)

// NewAsyncRequestWorkerWithFunc creates a worker using a function as the processor
// If retryConfig is nil, no retry logic will be used
func NewAsyncRequestWorkerWithFunc[T any, R any](
	bufferSize int,
	retryConfig *retry.RetryConfig,
	processFn func(ctx context.Context, req Request[T]) (R, error),
) *AsyncRequestProcessor[T, R] {
	processor := FunctionProcessor[T, R]{ProcessFn: processFn}
	return NewAsyncRequstProcessor(bufferSize, processor, retryConfig)
}

// Start begins the worker's processing loop in a separate goroutine
func (w *AsyncRequestProcessor[T, R]) Start() {
	w.wg.Add(1)
	go w.processLoop()
}

// Stop gracefully shuts down the worker after processing remaining requests
func (w *AsyncRequestProcessor[T, R]) Stop() {
	w.cancel()
	w.wg.Wait()
	close(w.responseChan)
}

// Submit sends a new request to the worker
// Returns false if the worker is unable to accept the request
func (w *AsyncRequestProcessor[T, R]) Submit(req Request[T]) bool {
	select {
	case <-w.ctx.Done():
		return false
	case w.requestChan <- req:
		return true
	default:
		// Channel is full
		return false
	}
}

// Responses returns the channel for receiving responses
func (w *AsyncRequestProcessor[T, R]) Responses() <-chan Response[R] {
	return w.responseChan
}

// processLoop is the main worker routine that processes requests synchronously
func (w *AsyncRequestProcessor[T, R]) processLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			// Process remaining items in the channel before exiting
			for {
				select {
				case req := <-w.requestChan:
					w.processRequest(req)
				default:
					return
				}
			}

		case req := <-w.requestChan:
			w.processRequest(req)
		}
	}
}

// processRequest handles processing a single request with retry logic
func (w *AsyncRequestProcessor[T, R]) processRequest(req Request[T]) {
	startTime := time.Now()

	var responseData R
	var err error

	// If no retry config is set, process the request directly
	if w.retryConfig == nil {
		responseData, err = w.process(req)
	} else {
		// Retry logic
		err = retry.RetryWithBackoff(w.ctx, *w.retryConfig, func(ctx context.Context) error {
			responseData, err = w.process(req)
			return err
		})
	}

	duration := time.Since(startTime)

	// Send the response back through the response channel
	select {
	case w.responseChan <- Response[R]{
		RequestID: req.ID,
		Data:      responseData,
		Error:     err,
		Duration:  duration,
	}:
	case <-w.ctx.Done():
		// Worker is shutting down, don't try to send results
	}
}

func (w *AsyncRequestProcessor[T, R]) process(req Request[T]) (R, error) {
	// Create a context for this specific request that inherits from the worker context
	reqCtx, cancel := context.WithTimeout(w.ctx, w.retryConfig.MaxDuration)

	// Process the request using the custom processor
	responseData, err := w.processor.Process(reqCtx, req)
	cancel() // Always cancel the request context

	if err == nil {
		return responseData, nil
	}

	return responseData, err
}
