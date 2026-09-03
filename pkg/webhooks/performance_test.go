package webhooks

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func BenchmarkDeliver(b *testing.B) {
	for _, size := range []int{0, 1024, 64 << 10, MaxPayloadBytes} {
		b.Run(sizeName(size), func(b *testing.B) {
			transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return response(http.StatusNoContent, "", ""), nil
			})
			config, err := validateConfig(Config{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: discardRecorder{}, Retry: RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, AttemptTimeout: time.Second})
			if err != nil {
				b.Fatal(err)
			}
			d := newDispatcher(config, dependencies{client: &http.Client{Transport: transport}, now: fixedClock, wait: noWait})
			msg := validMessage()
			msg.Body = make([]byte, size)
			b.ReportAllocs()
			b.SetBytes(int64(size))
			b.ResetTimer()
			for range b.N {
				if _, err := d.Deliver(context.Background(), msg); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("pathological-response-overflow", func(b *testing.B) {
		config, err := validateConfig(Config{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: discardRecorder{}, Retry: RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, AttemptTimeout: time.Second})
		if err != nil {
			b.Fatal(err)
		}
		responseBody := strings.Repeat("x", MaxResponseBytes+1)
		transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(responseBody))}, nil
		})
		d := newDispatcher(config, dependencies{client: &http.Client{Transport: transport}, now: fixedClock, wait: noWait})
		msg := validMessage()
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := d.Deliver(context.Background(), msg); err == nil {
				b.Fatal("overflow unexpectedly succeeded")
			}
		}
	})
}

func sizeName(size int) string {
	switch size {
	case 0:
		return "empty"
	case 1024:
		return "small-1KiB"
	case 64 << 10:
		return "typical-64KiB"
	case MaxPayloadBytes:
		return "boundary-1MiB"
	default:
		return "other"
	}
}
