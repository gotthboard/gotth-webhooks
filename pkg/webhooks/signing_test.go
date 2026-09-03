package webhooks

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestBuildRequestConformanceVector(t *testing.T) {
	t.Parallel()

	msg, err := validateMessage(Message{
		Endpoint:    "https://EXAMPLE.com/hook?b=2&a=%2F%3f&flag&x=one+two/three?four",
		DeliveryID:  "delivery-1",
		EventType:   "thing.changed",
		ContentType: "application/json",
		Body:        []byte("{}"),
	})
	if err != nil {
		t.Fatal(err)
	}
	secret := Secret{KeyID: "key-1", Value: []byte("0123456789abcdef0123456789abcdef")}
	req, canonical, err := buildRequest(context.Background(), msg, secret, 2, 1700000000)
	if err != nil {
		t.Fatal(err)
	}
	wantCanonical := "gotth-webhook-signature-v1\nPOST\nhttps://example.com:443/hook?b=2&a=%2F%3f&flag&x=one+two/three?four\ndelivery-1\n2\n1700000000\nthing.changed\napplication/json\nkey-1\n44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a\n"
	if string(canonical) != wantCanonical {
		t.Fatalf("canonical = %q, want %q", canonical, wantCanonical)
	}
	if got, want := req.Header.Get(HeaderSignature), "v1=5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f"; got != want {
		t.Fatalf("signature = %q, want %q", got, want)
	}
	if req.Method != "POST" || req.URL.String() != "https://example.com:443/hook?b=2&a=%2F%3f&flag&x=one+two/three?four" || req.URL.RawQuery != "b=2&a=%2F%3f&flag&x=one+two/three?four" || req.Host != "example.com:443" {
		t.Fatalf("unexpected request target: %s %s host=%s", req.Method, req.URL, req.Host)
	}
	if req.Header.Get(HeaderDeliveryID) != "delivery-1" || req.Header.Get(HeaderAttempt) != "2" || req.Header.Get(HeaderTimestamp) != "1700000000" || req.Header.Get(HeaderEvent) != "thing.changed" || req.Header.Get(HeaderKeyID) != "key-1" {
		t.Fatalf("missing signed metadata: %#v", req.Header)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "{}" {
		t.Fatalf("body = %q", body)
	}
}

func TestSignatureBindsEveryField(t *testing.T) {
	t.Parallel()

	base := Message{Endpoint: "https://example.com/hook", DeliveryID: "id-1", EventType: "event.one", ContentType: "application/json", Body: []byte("body")}
	secret := Secret{KeyID: "key-1", Value: []byte(strings.Repeat("s", minSecretBytes))}
	variants := []Message{
		{Endpoint: "https://example.org/hook", DeliveryID: "id-1", EventType: "event.one", ContentType: "application/json", Body: []byte("body")},
		{Endpoint: "https://example.com/other", DeliveryID: "id-1", EventType: "event.one", ContentType: "application/json", Body: []byte("body")},
		{Endpoint: "https://example.com/hook", DeliveryID: "id-2", EventType: "event.one", ContentType: "application/json", Body: []byte("body")},
		{Endpoint: "https://example.com/hook", DeliveryID: "id-1", EventType: "event.two", ContentType: "application/json", Body: []byte("body")},
		{Endpoint: "https://example.com/hook", DeliveryID: "id-1", EventType: "event.one", ContentType: "text/plain", Body: []byte("body")},
		{Endpoint: "https://example.com/hook", DeliveryID: "id-1", EventType: "event.one", ContentType: "application/json", Body: []byte("changed")},
	}
	signature := func(msg Message, attempt int, timestamp int64, key Secret) string {
		validated, err := validateMessage(msg)
		if err != nil {
			t.Fatal(err)
		}
		req, _, err := buildRequest(context.Background(), validated, key, attempt, timestamp)
		if err != nil {
			t.Fatal(err)
		}
		return req.Header.Get(HeaderSignature)
	}
	want := signature(base, 1, 1, secret)
	for i, variant := range variants {
		if got := signature(variant, 1, 1, secret); got == want {
			t.Errorf("variant %d did not change signature", i)
		}
	}
	if signature(base, 2, 1, secret) == want || signature(base, 1, 2, secret) == want {
		t.Error("attempt or timestamp did not change signature")
	}
	otherKeyID := secret
	otherKeyID.KeyID = "key-2"
	if signature(base, 1, 1, otherKeyID) == want {
		t.Error("key ID did not change signature")
	}
}

func TestSignatureNormalizesDefaultPortAndRejectsEmptyPort(t *testing.T) {
	t.Parallel()

	secret := Secret{KeyID: "key-1", Value: []byte("0123456789abcdef0123456789abcdef")}
	signature := func(endpoint string) (string, string) {
		t.Helper()
		msg, err := validateMessage(Message{
			Endpoint: endpoint, DeliveryID: "delivery-1", EventType: "thing.changed",
			ContentType: "application/json", Body: []byte("{}"),
		})
		if err != nil {
			t.Fatal(err)
		}
		req, canonical, err := buildRequest(context.Background(), msg, secret, 2, 1700000000)
		if err != nil {
			t.Fatal(err)
		}
		return string(canonical), req.Header.Get(HeaderSignature)
	}
	omittedCanonical, omittedSignature := signature("https://EXAMPLE.com/hook?q=1")
	explicitCanonical, explicitSignature := signature("https://example.com:443/hook?q=1")
	if omittedCanonical != explicitCanonical || omittedSignature != explicitSignature {
		t.Fatalf("default-port normalization diverged:\nomitted=%q %q\nexplicit=%q %q", omittedCanonical, omittedSignature, explicitCanonical, explicitSignature)
	}
	for _, endpoint := range []string{
		"https://example.com:/hook",
		"https://[2001:4860:4860::8888]:/hook",
	} {
		_, err := validateMessage(Message{
			Endpoint: endpoint, DeliveryID: "delivery-1", EventType: "thing.changed",
			ContentType: "application/json", Body: []byte("{}"),
		})
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("endpoint=%q error=%v", endpoint, err)
		}
	}
}
