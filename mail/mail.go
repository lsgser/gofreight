package mail

import (
	"sync"
)

// Message represents an email message.
type Message struct {
	To      []string
	From    string
	Subject string
	Body    string
	HTML    string
}

// Mailer sends email messages.
type Mailer interface {
	Send(msg Message) error
}

// LogMailer logs emails instead of sending (development/test).
type LogMailer struct {
	mu       sync.Mutex
	Messages []Message
}

// NewLogMailer creates a mailer that stores sent messages in memory.
func NewLogMailer() *LogMailer {
	return &LogMailer{}
}

func (m *LogMailer) Send(msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, msg)
	return nil
}

// Sent returns all sent messages.
func (m *LogMailer) Sent() []Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Message{}, m.Messages...)
}

// Last returns the most recently sent message.
func (m *LogMailer) Last() *Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Messages) == 0 {
		return nil
	}
	msg := m.Messages[len(m.Messages)-1]
	return &msg
}

// Reset clears sent messages (for tests).
func (m *LogMailer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = nil
}

// Count returns the number of sent messages.
func (m *LogMailer) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Messages)
}
