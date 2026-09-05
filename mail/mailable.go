package mail

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"path/filepath"
)

// Mailable is an email with a template view (Laravel Mailable).
type Mailable struct {
	To       []string
	From     string
	Subject  string
	View     string
	Data     map[string]any
	viewsDir string
}

// NewMailable creates a mailable with template path relative to viewsDir.
func NewMailable(viewsDir, view, subject string, to ...string) *Mailable {
	return &Mailable{
		viewsDir: viewsDir,
		View:     view,
		Subject:  subject,
		To:       to,
		Data:     make(map[string]any),
	}
}

// With sets template data.
func (m *Mailable) With(key string, value any) *Mailable {
	m.Data[key] = value
	return m
}

// Render builds the HTML body from the template file.
func (m *Mailable) Render() (string, error) {
	path := filepath.Join(m.viewsDir, m.View)
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New(m.View).Parse(string(raw))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, m.Data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Send renders and sends via the mailer.
func (m *Mailable) Send(mailer Mailer) error {
	html, err := m.Render()
	if err != nil {
		return err
	}
	return mailer.Send(Message{
		To:      m.To,
		From:    m.From,
		Subject: m.Subject,
		HTML:    html,
	})
}

// QueuedMailer wraps a mailer and dispatches sends via a job queue.
type QueuedMailer struct {
	Mailer Mailer
	Queue  JobDispatcher
}

// JobDispatcher dispatches background work.
type JobDispatcher interface {
	DispatchFunc(fn func(ctx context.Context) error)
}

// Send queues the message for background delivery.
func (q *QueuedMailer) Send(msg Message) error {
	msgCopy := msg
	q.Queue.DispatchFunc(func(ctx context.Context) error {
		return q.Mailer.Send(msgCopy)
	})
	return nil
}

// QueueMailable queues a mailable for delivery.
func QueueMailable(q *QueuedMailer, m *Mailable) error {
	mCopy := *m
	q.Queue.DispatchFunc(func(ctx context.Context) error {
		return mCopy.Send(q.Mailer)
	})
	return nil
}
