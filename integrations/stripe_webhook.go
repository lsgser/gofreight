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
	"net/url"
	"strings"
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

// Track sends an analytics event to the configured provider.
func (a *Analytics) Track(ctx context.Context, event string, props map[string]any) error {
	if !a.enabled {
		return nil
	}
	switch strings.ToLower(a.Provider) {
	case "posthog":
		return posthogTrack(ctx, a.APIKey, event, props)
	default:
		return segmentTrack(ctx, a.APIKey, event, props)
	}
}

func posthogTrack(ctx context.Context, apiKey, event string, props map[string]any) error {
	payload := map[string]any{
		"api_key": apiKey,
		"event":   event,
		"properties": props,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://app.posthog.com/capture/", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func segmentTrack(ctx context.Context, writeKey, event string, props map[string]any) error {
	payload := map[string]any{"event": event, "properties": props}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.segment.io/v1/track", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(writeKey, "")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// CreateCustomer creates a Stripe customer and returns the customer ID.
func (s *Stripe) CreateCustomer(ctx context.Context, email string) (string, error) {
	if !s.enabled {
		return "", ErrNotConfigured
	}
	form := url.Values{"email": {email}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.stripe.com/v1/customers", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.SecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

// Charge creates a Stripe payment intent (amount in cents).
func (s *Stripe) Charge(ctx context.Context, amount int64, currency, customerID string) (string, error) {
	if !s.enabled {
		return "", ErrNotConfigured
	}
	form := url.Values{
		"amount":   {fmt.Sprintf("%d", amount)},
		"currency": {currency},
		"customer": {customerID},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.stripe.com/v1/payment_intents", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(s.SecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}
