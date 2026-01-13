package retry_test

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/Pix4D/go-kit/retry"
)

func TestRetrySuccessOnFirstAttempt(t *testing.T) {
	var sleeps []time.Duration
	rtr := retry.Retry{
		UpTo:         5 * time.Second,
		FirstDelay:   1 * time.Second,
		BackoffLimit: 1 * time.Minute,
		Log:          makeLog(),
		SleepFn:      func(d time.Duration) { sleeps = append(sleeps, d) },
	}
	workFn := func() error { return nil }

	err := rtr.Do(retry.ConstantBackoff, retryOnError, workFn)
	if err != nil {
		t.Errorf("%s:\nhave: %v\nwant: %v", "retry.Do", err, "<no error>")
	}
	if have, want := sleeps, []time.Duration{}; slices.Compare(have, want) != 0 {
		t.Errorf("%s:\nhave: %v\nwant: %v", "sleeps", have, want)
	}
}

func TestRetrySuccessOnThirdAttempt(t *testing.T) {
	var sleeps []time.Duration
	rtr := retry.Retry{
		UpTo:         5 * time.Second,
		FirstDelay:   1 * time.Second,
		BackoffLimit: 1 * time.Minute,
		Log:          makeLog(),
		SleepFn:      func(d time.Duration) { sleeps = append(sleeps, d) },
	}
	attempt := 0
	workFn := func() error {
		attempt++
		if attempt == 3 {
			return nil
		}
		return fmt.Errorf("attempt %d", attempt)
	}

	err := rtr.Do(retry.ConstantBackoff, retryOnError, workFn)
	if err != nil {
		t.Errorf("%s:\nhave: %v\nwant: %v", "retry.Do", err, "<no error>")
	}
	wantSleeps := []time.Duration{time.Second, time.Second}
	if have, want := sleeps, wantSleeps; slices.Compare(have, want) != 0 {
		t.Errorf("%s:\nhave: %v\nwant: %v", "sleeps", have, want)
	}
}

func TestRetryFailureRunOutOfTime(t *testing.T) {
	var sleeps []time.Duration
	rtr := retry.Retry{
		UpTo:         5 * time.Second,
		FirstDelay:   1 * time.Second,
		BackoffLimit: 1 * time.Minute,
		Log:          makeLog(),
		SleepFn:      func(d time.Duration) { sleeps = append(sleeps, d) },
	}
	ErrAlwaysFail := errors.New("I always fail")
	workFn := func() error { return ErrAlwaysFail }

	err := rtr.Do(retry.ConstantBackoff, retryOnError, workFn)

	if have, want := err, ErrAlwaysFail; !errors.Is(have, want) {
		t.Errorf("%s:\nhave: %v\nwant: %v", "retry.Do", have, want)
	}
	wantSleeps := []time.Duration{
		time.Second, time.Second, time.Second, time.Second, time.Second,
	}
	if have, want := sleeps, wantSleeps; slices.Compare(have, want) != 0 {
		t.Errorf("%s:\nhave: %v\nwant: %v", "sleeps", have, want)
	}
}

func TestRetryExponentialBackOff(t *testing.T) {
	var sleeps []time.Duration
	rtr := retry.Retry{
		FirstDelay:   1 * time.Second,
		BackoffLimit: 4 * time.Second,
		UpTo:         11 * time.Second,
		SleepFn:      func(d time.Duration) { sleeps = append(sleeps, d) },
		Log:          makeLog(),
	}
	ErrAlwaysFail := errors.New("I always fail")
	workFn := func() error { return ErrAlwaysFail }

	err := rtr.Do(retry.ExponentialBackoff, retryOnError, workFn)

	if have, want := err, ErrAlwaysFail; !errors.Is(have, want) {
		t.Errorf("%s:\nhave: %v\nwant: %v", "retry.Do", have, want)
	}
	wantSleeps := []time.Duration{
		time.Second, 2 * time.Second, 4 * time.Second, 4 * time.Second,
	}
	if have, want := sleeps, wantSleeps; slices.Compare(have, want) != 0 {
		t.Errorf("%s:\nhave: %v\nwant: %v", "sleeps", have, want)
	}
}

func TestRetryFailureHardFailOnSecondAttempt(t *testing.T) {
	var sleeps []time.Duration
	rtr := retry.Retry{
		UpTo:         5 * time.Second,
		FirstDelay:   1 * time.Second,
		BackoffLimit: 1 * time.Minute,
		Log:          makeLog(),
		SleepFn:      func(d time.Duration) { sleeps = append(sleeps, d) },
	}
	ErrUnrecoverable := errors.New("I am unrecoverable")
	classifierFn := func(err error) retry.Action {
		if errors.Is(err, ErrUnrecoverable) {
			return retry.HardFail
		}
		if err != nil {
			return retry.SoftFail
		}
		return retry.Success
	}
	attempt := 0
	workFn := func() error {
		attempt++
		if attempt == 2 {
			return ErrUnrecoverable
		}
		return fmt.Errorf("attempt %d", attempt)
	}

	err := rtr.Do(retry.ConstantBackoff, classifierFn, workFn)

	if have, want := err, ErrUnrecoverable; !errors.Is(have, want) {
		t.Errorf("%s:\nhave: %v\nwant: %v", "retry.Do", have, want)
	}
	wantSleeps := []time.Duration{time.Second}
	if have, want := sleeps, wantSleeps; slices.Compare(have, want) != 0 {
		t.Errorf("%s:\nhave: %v\nwant: %v", "sleeps", have, want)
	}
}

func retryOnError(err error) retry.Action {
	if err != nil {
		return retry.SoftFail
	}
	return retry.Success
}

func makeLog() *slog.Logger {
	out := io.Discard
	if testing.Verbose() {
		out = os.Stdout
	}
	return slog.New(slog.NewTextHandler(out, nil))
}
