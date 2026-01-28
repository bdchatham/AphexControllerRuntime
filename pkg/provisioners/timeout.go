// Package provisioners provides standardized interfaces and utilities for resource provisioning.
package provisioners

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"github.com/bdchatham/AphexControllerRuntime/pkg/constants"
)

// TimeoutError represents a provisioning timeout with step information.
type TimeoutError struct {
	Step     string
	Duration time.Duration
	Timeout  time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("provisioning step %q timed out after %v (timeout: %v)", e.Step, e.Duration, e.Timeout)
}

// IsTimeoutError returns true if the error is a TimeoutError.
func IsTimeoutError(err error) bool {
	var timeoutErr *TimeoutError
	return errors.As(err, &timeoutErr)
}

// CancellationError represents a context cancellation during provisioning.
type CancellationError struct {
	Step   string
	Reason string
}

func (e *CancellationError) Error() string {
	return fmt.Sprintf("provisioning step %q canceled: %s", e.Step, e.Reason)
}

// IsCancellationError returns true if the error is a CancellationError.
func IsCancellationError(err error) bool {
	var cancelErr *CancellationError
	return errors.As(err, &cancelErr)
}

// TimeoutRunner executes provisioning steps with timeout and cancellation support.
type TimeoutRunner struct {
	timeout time.Duration
	logger  logr.Logger
}

// NewTimeoutRunner creates a new TimeoutRunner with the specified timeout.
// If timeout is 0, the default provisioning timeout is used.
func NewTimeoutRunner(timeout time.Duration, logger logr.Logger) *TimeoutRunner {
	if timeout == 0 {
		timeout = constants.DefaultProvisioningTimeout
	}
	return &TimeoutRunner{
		timeout: timeout,
		logger:  logger.WithName("TimeoutRunner"),
	}
}

// ProvisioningStep represents a single provisioning step with a name and function.
type ProvisioningStep struct {
	Name string
	Fn   func(ctx context.Context) error
}

// RunStep executes a single provisioning step with timeout.
// Returns a TimeoutError if the step exceeds the timeout.
// Returns a CancellationError if the context is canceled.
func (r *TimeoutRunner) RunStep(ctx context.Context, step ProvisioningStep) error {
	stepCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	r.logger.V(1).Info("Starting provisioning step",
		"step", step.Name,
		"timeout", r.timeout,
	)

	startTime := time.Now()

	errCh := make(chan error, 1)
	go func() {
		errCh <- step.Fn(stepCtx)
	}()

	select {
	case err := <-errCh:
		duration := time.Since(startTime)
		if err != nil {
			r.logger.Error(err, "Provisioning step failed",
				"step", step.Name,
				"duration", duration,
			)
			return fmt.Errorf("step %q failed: %w", step.Name, err)
		}
		r.logger.V(1).Info("Provisioning step completed",
			"step", step.Name,
			"duration", duration,
		)
		return nil

	case <-stepCtx.Done():
		duration := time.Since(startTime)
		if ctx.Err() == context.Canceled {
			r.logger.Info("Provisioning step canceled",
				"step", step.Name,
				"duration", duration,
			)
			return &CancellationError{
				Step:   step.Name,
				Reason: "parent context canceled",
			}
		}
		r.logger.Info("Provisioning step timed out",
			"step", step.Name,
			"duration", duration,
			"timeout", r.timeout,
		)
		return &TimeoutError{
			Step:     step.Name,
			Duration: duration,
			Timeout:  r.timeout,
		}
	}
}

// RunSteps executes multiple provisioning steps sequentially with timeout.
// Checks for context cancellation between steps.
// Returns on the first error encountered.
func (r *TimeoutRunner) RunSteps(ctx context.Context, steps []ProvisioningStep) error {
	for i, step := range steps {
		if err := CheckContextCancellation(ctx, step.Name); err != nil {
			return err
		}

		r.logger.V(1).Info("Executing provisioning step",
			"step", i+1,
			"totalSteps", len(steps),
			"stepName", step.Name,
		)

		if err := r.RunStep(ctx, step); err != nil {
			return err
		}
	}

	return nil
}

// CheckContextCancellation checks if the context has been canceled.
// Returns a CancellationError if canceled, nil otherwise.
func CheckContextCancellation(ctx context.Context, stepName string) error {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.Canceled {
			return &CancellationError{
				Step:   stepName,
				Reason: "context canceled before step started",
			}
		}
		if ctx.Err() == context.DeadlineExceeded {
			return &CancellationError{
				Step:   stepName,
				Reason: "context deadline exceeded before step started",
			}
		}
		return &CancellationError{
			Step:   stepName,
			Reason: ctx.Err().Error(),
		}
	default:
		return nil
	}
}

// WithTimeout wraps a context with the default provisioning timeout.
func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, constants.DefaultProvisioningTimeout)
}

// WithCustomTimeout wraps a context with a custom timeout.
func WithCustomTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
