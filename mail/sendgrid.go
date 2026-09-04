package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SendGridMailer sends email via SendGrid API.
type SendGridMailer struct {
	APIKey string
	From   string
}

func NewSendGrid(apiKey, from string) *SendGridMailer {
	return &SendGridMailer{APIKey: apiKey, From: from}
}

func (m *SendGridMailer) Send(msg Message) error {
	from := msg.From
	if from == "" {
		from = m.From
	}

	contentType := "text/plain"
	content := msg.Body
	if msg.HTML != "" {
		contentType = "text/html"
		content = msg.HTML
	}

	payload := map[string]any{
		"personalizations": []map[string]any{
			{"to": toRecipients(msg.To)},
		},
		"from":    map[string]string{"email": from},
		"subject": msg.Subject,
		"content": []map[string]string{
			{"type": contentType, "value": content},
		},
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid: status %d", resp.StatusCode)
	}
	return nil
}

func toRecipients(addrs []string) []map[string]string {
	var out []map[string]string
	for _, a := range addrs {
		out = append(out, map[string]string{"email": a})
	}
	return out
}
