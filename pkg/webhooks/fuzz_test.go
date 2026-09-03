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

func FuzzCanonicalContentType(f *testing.F) {
	for _, seed := range []string{
		"application/json",
		` Application/JSON ; note="hello world" `,
		`text/plain; title="café"`,
		"application/json\t",
		`application/json; note="a` + "\x00" + `b"`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		canonical, err := canonicalContentType(value)
		containsControl := false
		for i := range len(value) {
			if value[i] < 0x20 || value[i] == 0x7f {
				containsControl = true
				break
			}
		}
		if containsControl && err == nil {
			t.Fatalf("accepted ASCII control in %q as %q", value, canonical)
		}
		if err != nil {
			return
		}
		for i := range len(canonical) {
			if canonical[i] < 0x20 || canonical[i] == 0x7f {
				t.Fatalf("canonical output contains ASCII control: input=%q output=%q", value, canonical)
			}
		}
		repeated, err := canonicalContentType(canonical)
		if err != nil || repeated != canonical {
			t.Fatalf("canonical form is not idempotent: input=%q canonical=%q repeated=%q error=%v", value, canonical, repeated, err)
		}
	})
}
