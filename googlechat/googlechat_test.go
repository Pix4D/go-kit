package googlechat_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/marco-m/rosina/assert"

	"github.com/Pix4D/go-kit/googlechat"
	"github.com/Pix4D/go-kit/internal/testutils"
)

func TestTextMessageIntegration(t *testing.T) {
	gchatUrl := os.Getenv("GOKIT_TEST_GCHAT_WEBHOOK")
	if gchatUrl == "" {
		t.Skip("Skipping integration test. See CONTRIBUTING for how to enable.")
	}

	log := testutils.MakeTestLog()
	ts := time.Now().Format("2006-01-02 15:04:05 MST")
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}
	threadKey := "banana-" + user
	text := fmt.Sprintf("%s message oink! 🐷 sent to thread %s by user %s",
		ts, threadKey, user)

	reply, err := googlechat.TextMessage(log, googlechat.DefaultRetry(log),
		googlechat.DefaultTimeout, gchatUrl, threadKey, text)

	assert.NoError(t, err, "TextMessage")
	assert.Contains(t, reply.Text, text, "TextMessage reply")
}

func TestTextMessageRetryDueToStatusCodeAndPass(t *testing.T) {
	log := testutils.MakeTestLog()
	var sleepsCountSpy int
	rtr := googlechat.DefaultRetry(log)
	rtr.SleepFn = func(d time.Duration) { sleepsCountSpy++ }

	test := func(codes []int, wantSleeps int) {
		t.Helper()
		sleepsCountSpy = 0
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if len(codes) == 0 {
					t.Fatalf("fake server: no more status codes left")
				}
				var code int
				code, codes = codes[0], codes[1:]
				w.WriteHeader(code)
				w.Write([]byte("{}")) //nolint:errcheck
			}))
		defer ts.Close()

		_, err := googlechat.TextMessage(log, rtr, googlechat.DefaultTimeout, ts.URL,
			"key", "bananas are ripe")

		assert.NoError(t, err, "TextMessage")
		assert.Equal(t, sleepsCountSpy, wantSleeps, "sleeps")
	}

	test([]int{http.StatusOK}, 0)
	test([]int{http.StatusTooManyRequests, http.StatusOK}, 1)
	test([]int{http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusOK}, 2)
}

func TestTextMessageRetryDueToStatusCodeAndFail(t *testing.T) {
	log := testutils.MakeTestLog()
	var sleepTimeSpy time.Duration
	rtr := googlechat.DefaultRetry(log)
	rtr.SleepFn = func(d time.Duration) { sleepTimeSpy += d }

	test := func(code int, wantSlept time.Duration) {
		t.Helper()
		sleepTimeSpy = 0
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(code)
			}))
		defer ts.Close()

		_, err := googlechat.TextMessage(log, rtr, googlechat.DefaultTimeout, ts.URL,
			"key", "bananas are ripe")

		assert.ErrorContains(t, err, http.StatusText(code), "TextMessage")
		assert.Equal(t, sleepTimeSpy, wantSlept, "sleeps")
	}

	test(http.StatusForbidden, 0)                 // not retriable: fails immediately.
	test(http.StatusTooManyRequests, rtr.UpTo)    // retriable; fails after consuming all retries.
	test(http.StatusServiceUnavailable, rtr.UpTo) // retriable; fails after consuming all retries.
}

func TestTextMessageRetryDueToRequestTimeout(t *testing.T) {
	log := testutils.MakeTestLog()
	var sleepsCountSpy int
	rtr := googlechat.DefaultRetry(log)
	rtr.SleepFn = func(d time.Duration) { sleepsCountSpy++ }
	const timeout = 10 * time.Millisecond

	test := func(failingReqs, wantSleeps int) {
		t.Helper()
		sleepsCountSpy = 0
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if failingReqs > 0 {
					failingReqs--
					time.Sleep(10 * timeout)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}")) //nolint:errcheck
			}))
		defer ts.Close()

		_, err := googlechat.TextMessage(log, rtr, timeout, ts.URL, "key", "bananas")

		assert.NoError(t, err, "TextMessage")
		assert.Equal(t, sleepsCountSpy, wantSleeps, "sleeps")
	}

	test(0, 0)
	test(1, 1)
	test(5, 5)
}

func TestRedactURL(t *testing.T) {
	hook := "https://chat.googleapis.com/v1/spaces/SSS/messages?key=KKK&token=TTT"
	want := "https://chat.googleapis.com/v1/spaces/SSS/messages?REDACTED"
	theURL, err := url.Parse(hook)
	assert.NoError(t, err, "url.Parse")

	have := googlechat.RedactURL(theURL).String()

	assert.Equal(t, have, want, "RedactURL")
}

func TestRedactString(t *testing.T) {
	hook := "https://chat.googleapis.com/v1/spaces/SSS/messages?key=KKK&token=TTT"
	want := "https://chat.googleapis.com/v1/spaces/SSS/messages?REDACTED"

	have := googlechat.RedactURLString(hook)

	assert.Equal(t, have, want, "RedactURLString")
}
