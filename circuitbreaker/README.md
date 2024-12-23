# Circuit Breaker Pattern

A thread-safe implementation of the Circuit Breaker pattern in Go, designed to prevent cascading failures in distributed systems.

## Overview

The Circuit Breaker pattern manages failures by temporarily stopping operations that are likely to fail. Like an electrical circuit breaker, it prevents system overload and allows time for problems to be resolved.

## States

- **Closed**: Normal operation mode
- **Open**: Failure mode, fast-fails requests
- **Half-Open**: Recovery mode, allows test requests

## Features

- Thread-safe operations
- Configurable failure threshold
- Adjustable reset timeout
- State change notifications
- Support for concurrent requests
- Automatic recovery attempts

## Usage

```go
cb := circuitbreaker.New(circuitbreaker.Options{
    FailureThreshold: 3,
    ResetTimeout:     10 * time.Second,
    OnStateChange: func(from, to State) {
        log.Printf("Circuit state changed from %s to %s", from, to)
    },
})

// Make a service call
err := cb.Execute(func() error {
    return makeExternalServiceCall()
})

if err != nil {
    // Handle error
}
```

## Configuration

- `FailureThreshold`: Number of consecutive failures before opening the circuit
- `ResetTimeout`: Duration to wait before attempting recovery
- `OnStateChange`: Callback function for state transition notifications

## State Transitions

1. **Closed → Open**: Occurs when failures reach the threshold
2. **Open → Half-Open**: Occurs after the reset timeout
3. **Half-Open → Closed**: Occurs after successful test requests
4. **Half-Open → Open**: Occurs if a test request fails
