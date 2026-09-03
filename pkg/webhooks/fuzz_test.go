package webhooks

import (
	"context"
	"strings"
	"testing"
)

func FuzzParseEndpoint(f *testing.F) {
	for _, seed := range []string{
		"https://example.com/hook",
		"https://example.com/a?b=1",
		"http://127.0.0.1/",
		"https://user@example.com/",
		"\x00",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		ep, err := parseEndpoint(raw)
		if err != nil {
			return
		}
		if !strings.HasPrefix(ep.canonical, "https://") || ep.url.Scheme != "https" || ep.url.Port() != "443" || ep.url.User != nil || ep.url.Fragment != "" {
			t.Fatalf("unsafe accepted endpoint: raw=%q canonical=%q", raw, ep.canonical)
		}
	})
}

func FuzzSignedRequestDeterministic(f *testing.F) {
	f.Add([]byte(""), "event.one")
	f.Add([]byte("body"), "event-two")
	f.Fuzz(func(t *testing.T, body []byte, event string) {
		if len(body) > MaxPayloadBytes || validateToken(event, "event") != nil {
			return
		}
		msg, err := validateMessage(Message{Endpoint: "https://example.com/hook?q=1", DeliveryID: "delivery", EventType: event, ContentType: "application/octet-stream", Body: body})
		if err != nil {
			t.Fatal(err)
		}
		secret := Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}
		reqA, canonicalA, err := buildRequest(context.Background(), msg, secret, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		reqB, canonicalB, err := buildRequest(context.Background(), msg, secret, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if string(canonicalA) != string(canonicalB) || reqA.Header.Get(HeaderSignature) != reqB.Header.Get(HeaderSignature) {
			t.Fatal("canonical signing was nondeterministic")
		}
	})
}
