// Package retry implements a generic and customizable retry mechanism.
//
// Took some inspiration from:
// - https://github.com/eapache/go-resiliency/tree/main/retrier
package retry

import (
	"errors"
	"log/slog"
	"time"
)

// Action is returned by a WorkFunc to indicate to Retry how to proceed.
type Action int

const (
	// Success informs Retry that the attempt has been a success.
	Success Action = iota
	// HardFail informs Retry that the attempt has been a hard failure and
	// thus should abort retrying.
	HardFail
	// SoftFail informs Retry that the attempt has been a soft failure and
	// thus should keep retrying.
	SoftFail
)

// Retry is the controller of the retry mechanism.
// See the examples in file retry_example_test.go.
type Retry struct {
	UpTo         time.Duration // Total maximum duration of the retries.
	FirstDelay   time.Duration // Duration of the first backoff.
	BackoffLimit time.Duration // Upper bound duration of a backoff.
	Log          *slog.Logger
	SleepFn      func(d time.Duration) // Optional; used only to override in tests.
}

// BackoffFunc returns the next backoff duration; called by [Retry.Do].
// You can use one of the ready-made functions [ConstantBackoff],
// [ExponentialBackoff] or write your own.
// Parameter err allows to optionally inspect the error that caused the retry
// and return a custom delay; this can be used in special cases such as when
// rate-limited with a fixed window; for an example see
// [github.com/Pix4D/go-kit/github.Backoff].
type BackoffFunc func(first bool, previous, limit time.Duration, err error) time.Duration

// WorkFunc does the unit of work that might fail and need to be retried; it also
// decides wether to proceed or not via the returned [Action]. Called by [Retry.Do].
// When the returned [Action] is [Success] or [HardFail], then the returned error,
// if any, will also be returned by [Retry.Do].
//
// For an example, see
//
//   - [github.CommitStatus.Add]
//   - retry_example_test.go
type WorkFunc func() (Action, error)

// Do is the loop of [Retry].
// See the examples in file retry_example_test.go.
func (rtr Retry) Do(
	backoffFn BackoffFunc,
	workFn WorkFunc,
) error {
	if rtr.FirstDelay <= 0 {
		return errors.New("FirstDelay must be positive")
	}
	if rtr.BackoffLimit <= 0 {
		return errors.New("BackoffLimit must be positive")
	}
	if rtr.SleepFn == nil {
		rtr.SleepFn = time.Sleep
	}
	rtr.Log = rtr.Log.With("system", "retry") // FIXME maybe better constructor???

	delay := rtr.FirstDelay
	totalDelay := 0 * time.Second

	for attempt := 1; ; attempt++ {
		action, err := workFn()
		switch action {
		case Success:
			rtr.Log.Info("success", "attempt", attempt, "totalDelay", totalDelay)
			return err
		case HardFail:
			return err
		case SoftFail:
			delay = backoffFn(attempt == 1, delay, rtr.BackoffLimit, err)
			totalDelay += delay
			if totalDelay > rtr.UpTo {
				rtr.Log.Error("would wait for too long", "attempt", attempt,
					"delay", delay, "totalDelay", totalDelay, "UpTo", rtr.UpTo)
				return err
			}
			rtr.Log.Info("waiting", "attempt", attempt, "delay", delay,
				"totalDelay", totalDelay)
			rtr.SleepFn(delay)
		default:
			return errors.New("retry: internal error, please report")
		}
	}
}
