package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker is an interface defining the methods of the circuit breaker.
type CircuitBreaker interface {
	Execute(operation func() error) error
	GetState() State
}

// circuitBreaker implements the circuit breaker pattern
type circuitBreaker struct {
	mu sync.RWMutex

	failureThreshold int
	resetTimeout     time.Duration
	currentState     State
	failureCount     int
	lastFailureTime  time.Time
	successCount     int

	onStateChange func(from, to State)
	onError       func(err error)
}

// Options configures the circuit breaker
type Options struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	OnStateChange    func(from, to State)
	OnError          func(err error)
}

// New creates a new circuit breaker with the given options
func New(options Options) *circuitBreaker {
	if options.FailureThreshold <= 0 {
		options.FailureThreshold = 5
	}
	if options.ResetTimeout <= 0 {
		options.ResetTimeout = 60 * time.Second
	}
	if options.OnStateChange == nil {
		options.OnStateChange = func(from, to State) {}
	}
	if options.OnError == nil {
		options.OnError = func(err error) {}
	}

	return &circuitBreaker{
		failureThreshold: options.FailureThreshold,
		resetTimeout:     options.ResetTimeout,
		onStateChange:    options.OnStateChange,
		onError:          options.OnError,
		currentState:     StateClosed,
	}
}

// Execute runs the given function if the circuit breaker allows it
func (cb *circuitBreaker) Execute(operation func() error) error {
	if !cb.allowRequest() {
		return errors.New("circuit breaker is open")
	}

	err := operation()
	cb.handleResult(err)
	return err
}

func (cb *circuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.currentState {
	case StateClosed:
		return true
	case StateHalfOpen:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.toHalfOpen()
			cb.mu.RLock()
			return true
		}
		return false
	default:
		return false
	}
}

func (cb *circuitBreaker) handleResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure(err)
	} else {
		cb.onSuccess()
	}
}

func (cb *circuitBreaker) onSuccess() {
	switch cb.currentState {
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= 2 {
			cb.toState(StateClosed)
		}
	case StateClosed:
		cb.failureCount = 0
	}
}

func (cb *circuitBreaker) onFailure(err error) {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.currentState == StateClosed && cb.failureCount >= cb.failureThreshold {
		cb.toState(StateOpen)
	} else if cb.currentState == StateHalfOpen {
		cb.toState(StateOpen)
	}

	cb.onError(err)
}

func (cb *circuitBreaker) toHalfOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.toState(StateHalfOpen)
}

func (cb *circuitBreaker) toState(newState State) {
	if cb.currentState == newState {
		return
	}

	oldState := cb.currentState
	cb.currentState = newState
	cb.failureCount = 0
	cb.successCount = 0

	cb.onStateChange(oldState, newState)
}

// GetState returns the current state of the circuit breaker
func (cb *circuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.currentState
}

// Example usage:
func Example() {
	// Create a new circuit breaker
	cb := New(Options{
		FailureThreshold: 3,
		ResetTimeout:     10 * time.Second,
		OnStateChange: func(from, to State) {
			// Log state changes
		},
	})

	// Example service call
	err := cb.Execute(func() error {
		// Make external service call here
		return nil
	})

	if err != nil {
		// Handle error
	}
}

// Helper function to convert State to string for logging
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}
