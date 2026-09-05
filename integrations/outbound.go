package integrations

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Deliver sends a signed webhook POST request.
func (w *Webhook) Deliver(ctx context.Context, targetURL string, payload map[string]any) error {
	if !w.enabled {
		return ErrNotConfigured
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	sig := signPayload(body, w.Secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gofreight-Signature", sig)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook delivery failed: %d %s", resp.StatusCode, string(b))
	}
	return nil
}

func signPayload(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
