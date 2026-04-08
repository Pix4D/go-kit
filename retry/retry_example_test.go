package retry_test

// Go testable examples.
// Any function in a test package with prefix "Example" is a "testable example".
// It will be run as a test and the output must match the "Output:" in the comment.
// See https://go.dev/blog/examples.

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Pix4D/go-kit/retry"
)

func ExampleRetry() {
	rtr := retry.Retry{
		UpTo:         5 * time.Second,
		FirstDelay:   1 * time.Second,
		BackoffLimit: 1 * time.Second,
		Log:          makeExampleLog(),
	}

	workFn := func() (retry.Action, error) {
		err := func() error {
			// Do work. If something fails, as usual, return error.
			// ...
			return nil
		}()
		if err != nil {
			return retry.SoftFail, err
		}
		return retry.Success, err
	}

	err := rtr.Do(retry.ConstantBackoff, workFn)
	if err != nil {
		// Handle error...
		fmt.Println("error:", err)
	}

	// Output:
	// level=INFO msg=success system=retry attempt=1 totalDelay=0s
}

// Used in [ExampleRetry_customError].
var ErrBananaUnavailable = errors.New("banana service unavailable")

// Embedded in [BananaResponseError].
type BananaResponse struct {
	Amount int
	// In practice, more fields here...
}

// Used in [ExampleRetry_customError].
type BananaResponseError struct {
	Response *BananaResponse
	// In practice, more fields here...
}

func (eb BananaResponseError) Error() string {
	return "look at my fields, there is more information there"
}

func ExampleRetry_customError() {
	rtr := retry.Retry{
		UpTo:         30 * time.Second,
		FirstDelay:   2 * time.Second,
		BackoffLimit: 1 * time.Minute,
		Log:          makeExampleLog(),
		SleepFn:      func(d time.Duration) {}, // Only for the test!
	}

	attempt := 0
	workFn := func() (retry.Action, error) {
		err := func() error {
			attempt++
			if attempt == 3 {
				// Error wrapping is optional; we do it to show that it works also.
				return fmt.Errorf("workFn: %w",
					BananaResponseError{Response: &BananaResponse{Amount: 42}})
			}
			if attempt < 5 {
				return ErrBananaUnavailable
			}
			// On 5th attempt we finally succeed.
			return nil
		}()
		if err != nil {
			if brErr, ok := errors.AsType[BananaResponseError](err); ok {
				if brErr.Response.Amount == 42 {
					return retry.SoftFail, err
				}
				return retry.HardFail, err
			}
			if errors.Is(err, ErrBananaUnavailable) {
				return retry.SoftFail, err
			}
			return retry.HardFail, err
		}
		return retry.Success, err
	}

	err := rtr.Do(retry.ExponentialBackoff, workFn)
	if err != nil {
		// Handle error...
		fmt.Println("error:", err)
	}

	// Output:
	// level=INFO msg=waiting system=retry attempt=1 delay=2s totalDelay=2s
	// level=INFO msg=waiting system=retry attempt=2 delay=4s totalDelay=6s
	// level=INFO msg=waiting system=retry attempt=3 delay=8s totalDelay=14s
	// level=INFO msg=waiting system=retry attempt=4 delay=16s totalDelay=30s
	// level=INFO msg=success system=retry attempt=5 totalDelay=30s
}

func makeExampleLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{ReplaceAttr: removeTime}))
}

// removeTime removes time-dependent attributes from log/slog records, making
// the output of testable examples [1] deterministic.
// [1]: https://go.dev/blog/examples
func removeTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	// if a.Key == "elapsed" {
	//	return slog.Attr{}
	// }
	return a
}
