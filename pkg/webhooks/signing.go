package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
)

const signatureDomain = "gotth-webhook-signature-v1"

// buildRequest constructs one independently replayable signed POST. The
// returned canonical bytes exist for conformance testing and are not retained
// by Dispatcher. CPU time is Theta(1+b+m)+Htime(k); auxiliary space is
// Theta(1+m)+Hspace(k). b, m, and k are body, canonical metadata, and secret-key
// bytes; Htime/Hspace are delegated HMAC key-initialization costs. Hashing
// streams over the already-owned body; other delegated costs include SHA-256,
// URL/request construction, and body-reader allocation.
func buildRequest(ctx context.Context, msg validatedMessage, secret Secret, attempt int, timestamp int64) (*http.Request, []byte, error) {
	bodyDigest := sha256.Sum256(msg.body)
	canonical := []byte(fmt.Sprintf(
		"%s\nPOST\n%s\n%s\n%d\n%d\n%s\n%s\n%s\n%s\n",
		signatureDomain,
		msg.endpoint.canonical,
		msg.deliveryID,
		attempt,
		timestamp,
		msg.eventType,
		msg.contentType,
		secret.KeyID,
		hex.EncodeToString(bodyDigest[:]),
	))
	mac := hmac.New(sha256.New, secret.Value)
	_, _ = mac.Write(canonical)
	signature := "v1=" + hex.EncodeToString(mac.Sum(nil))

	urlCopy := *msg.endpoint.url
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlCopy.String(), bytes.NewReader(msg.body))
	if err != nil {
		return nil, nil, fmt.Errorf("build webhook request: %w", err)
	}
	req.Host = urlCopy.Host
	req.Header.Set("Content-Type", msg.contentType)
	req.Header.Set(HeaderDeliveryID, msg.deliveryID)
	req.Header.Set(HeaderEvent, msg.eventType)
	req.Header.Set(HeaderAttempt, strconv.Itoa(attempt))
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(timestamp, 10))
	req.Header.Set(HeaderKeyID, secret.KeyID)
	req.Header.Set(HeaderSignature, signature)
	return req, canonical, nil
}
