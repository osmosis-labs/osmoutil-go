package circuitbreaker_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	cb "github.com/osmosis-labs/osmoutil-go/circuitbreaker"
	"github.com/stretchr/testify/require"
)

const (
	defaultThreshold  = 3
	defaultTimeout    = 100 * time.Millisecond
	defaultWaitTime   = 200 * time.Millisecond
	testError         = "test error"
	circuitOpenError  = "circuit breaker is open"
	concurrentWorkers = 10
	iterationsPerTest = 100
)

func newTestCircuitBreaker(t *testing.T, opts ...func(*cb.Options)) cb.CircuitBreaker {
	options := cb.Options{
		FailureThreshold: defaultThreshold,
		ResetTimeout:     defaultTimeout,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return cb.New(options)
}

func TestCircuitBreaker(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "initial state is closed",
			test: testInitialState,
		},
		{
			name: "default options are valid",
			test: testDefaultOptions,
		},
		{
			name: "successful execution keeps circuit closed",
			test: testSuccessfulExecution,
		},
		{
			name: "failures open circuit",
			test: testFailureThreshold,
		},
		{
			name: "circuit transitions to half-open after timeout",
			test: testHalfOpenState,
		},
		{
			name: "successful recovery closes circuit",
			test: testSuccessfulRecovery,
		},
		{
			name: "state changes trigger callbacks",
			test: testStateChangeCallback,
		},
		{
			name: "handles concurrent executions",
			test: testConcurrentExecutions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func testInitialState(t *testing.T) {
	circuitBreaker := newTestCircuitBreaker(t)
	require.Equal(t, cb.StateClosed, circuitBreaker.GetState())
}

func testDefaultOptions(t *testing.T) {
	tests := []struct {
		name           string
		optionModifier func(*cb.Options)
		expectValid    bool
	}{
		{
			name: "zero threshold",
			optionModifier: func(o *cb.Options) {
				o.FailureThreshold = 0
			},
			expectValid: true, // should be corrected to default
		},
		{
			name: "zero timeout",
			optionModifier: func(o *cb.Options) {
				o.ResetTimeout = 0
			},
			expectValid: true, // should be corrected to default
		},
		{
			name: "negative values",
			optionModifier: func(o *cb.Options) {
				o.FailureThreshold = -1
				o.ResetTimeout = -1 * time.Second
			},
			expectValid: true, // should be corrected to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circuitBreaker := newTestCircuitBreaker(t, tt.optionModifier)
			require.Equal(t, cb.StateClosed, circuitBreaker.GetState())
		})
	}
}

func testSuccessfulExecution(t *testing.T) {
	circuitBreaker := newTestCircuitBreaker(t)

	err := circuitBreaker.Execute(func() error {
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, cb.StateClosed, circuitBreaker.GetState())
}

func testFailureThreshold(t *testing.T) {
	tests := []struct {
		name          string
		numFailures   int
		expectedState cb.State
	}{
		{
			name:          "below threshold",
			numFailures:   defaultThreshold - 1,
			expectedState: cb.StateClosed,
		},
		{
			name:          "at threshold",
			numFailures:   defaultThreshold,
			expectedState: cb.StateOpen,
		},
		{
			name:          "above threshold",
			numFailures:   defaultThreshold + 1,
			expectedState: cb.StateOpen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circuitBreaker := newTestCircuitBreaker(t)

			for i := 0; i < tt.numFailures; i++ {
				err := circuitBreaker.Execute(func() error {
					return errors.New(testError)
				})
				require.Error(t, err)
			}

			require.Equal(t, tt.expectedState, circuitBreaker.GetState())

			if tt.expectedState == cb.StateOpen {
				err := circuitBreaker.Execute(func() error {
					t.Error("This function should not be executed")
					return nil
				})
				require.EqualError(t, err, circuitOpenError)
			}
		})
	}
}

func testHalfOpenState(t *testing.T) {
	circuitBreaker := newTestCircuitBreaker(t)

	// Open the circuit
	for i := 0; i < defaultThreshold; i++ {
		_ = circuitBreaker.Execute(func() error {
			return errors.New(testError)
		})
	}

	require.Equal(t, cb.StateOpen, circuitBreaker.GetState())

	// Wait for reset timeout
	time.Sleep(defaultWaitTime)

	// First execution should be allowed (half-open state)
	err := circuitBreaker.Execute(func() error {
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, cb.StateHalfOpen, circuitBreaker.GetState())
}

func testSuccessfulRecovery(t *testing.T) {
	tests := []struct {
		name            string
		successfulCalls int
		expectedState   cb.State
	}{
		{
			name:            "single success in half-open",
			successfulCalls: 1,
			expectedState:   cb.StateHalfOpen,
		},
		{
			name:            "two successes close circuit",
			successfulCalls: 2,
			expectedState:   cb.StateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circuitBreaker := newTestCircuitBreaker(t)

			// Open the circuit
			for i := 0; i < defaultThreshold; i++ {
				_ = circuitBreaker.Execute(func() error {
					return errors.New(testError)
				})
			}

			// Wait for reset timeout
			time.Sleep(defaultWaitTime)

			// Execute successful calls
			for i := 0; i < tt.successfulCalls; i++ {
				err := circuitBreaker.Execute(func() error {
					return nil
				})
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedState, circuitBreaker.GetState())
		})
	}
}

func testStateChangeCallback(t *testing.T) {
	type stateChange struct {
		from cb.State
		to   cb.State
	}

	var stateChanges []stateChange
	var mu sync.Mutex

	circuitBreaker := newTestCircuitBreaker(t, func(o *cb.Options) {
		o.OnStateChange = func(from, to cb.State) {
			mu.Lock()
			stateChanges = append(stateChanges, stateChange{from, to})
			mu.Unlock()
		}
	})

	// Trigger state changes
	for i := 0; i < defaultThreshold; i++ {
		_ = circuitBreaker.Execute(func() error {
			return errors.New(testError)
		})
	}

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, stateChanges, 1)
	require.Equal(t, cb.StateClosed, stateChanges[0].from)
	require.Equal(t, cb.StateOpen, stateChanges[0].to)
}

func testConcurrentExecutions(t *testing.T) {
	circuitBreaker := newTestCircuitBreaker(t)
	var wg sync.WaitGroup

	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerTest; j++ {
				err := circuitBreaker.Execute(func() error {
					time.Sleep(time.Millisecond)
					return nil
				})
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		}()
	}

	wg.Wait()
	require.Equal(t, cb.StateClosed, circuitBreaker.GetState())
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state cb.State
		want  string
	}{
		{cb.StateClosed, "CLOSED"},
		{cb.StateHalfOpen, "HALF_OPEN"},
		{cb.StateOpen, "OPEN"},
		{cb.State(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			require.Equal(t, tt.want, tt.state.String())
		})
	}
}
