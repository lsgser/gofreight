package notification

import (
	"context"

	"github.com/lsgser/gofreight/mail"
)

// Notifiable receives notifications on one or more channels.
type Notifiable interface {
	NotificationChannels() []string
}

// Notification is a multi-channel message (Laravel Notifications).
type Notification interface {
	Via() []string
	ToMail(ctx context.Context) *mail.Message
	ToDatabase(ctx context.Context) map[string]any
}

// BaseNotification provides defaults for optional channels.
type BaseNotification struct {
	Channels []string
	MailMsg  *mail.Message
	DBData   map[string]any
}

func (n BaseNotification) Via() []string { return n.Channels }
func (n BaseNotification) ToMail(ctx context.Context) *mail.Message {
	_ = ctx
	return n.MailMsg
}
func (n BaseNotification) ToDatabase(ctx context.Context) map[string]any {
	_ = ctx
	return n.DBData
}

// Sender delivers notifications.
type Sender struct {
	Mailer mail.Mailer
	Store  DatabaseStore
}

// DatabaseStore persists database notifications.
type DatabaseStore interface {
	Store(userID int64, data map[string]any) error
}

// MemoryStore is an in-memory notification store (development).
type MemoryStore struct {
	items []storedNotification
}

type storedNotification struct {
	UserID int64
	Data   map[string]any
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (m *MemoryStore) Store(userID int64, data map[string]any) error {
	m.items = append(m.items, storedNotification{UserID: userID, Data: data})
	return nil
}

func (m *MemoryStore) All(userID int64) []map[string]any {
	var out []map[string]any
	for _, item := range m.items {
		if item.UserID == userID {
			out = append(out, item.Data)
		}
	}
	return out
}

// Send delivers a notification to a user across configured channels.
func (s *Sender) Send(ctx context.Context, userID int64, n Notification) error {
	if s == nil || n == nil {
		return nil
	}
	for _, channel := range n.Via() {
		switch channel {
		case "mail", "email":
			if s.Mailer != nil {
				if msg := n.ToMail(ctx); msg != nil {
					if err := s.Mailer.Send(*msg); err != nil {
						return err
					}
				}
			}
		case "database", "db":
			if s.Store != nil {
				if data := n.ToDatabase(ctx); data != nil {
					if err := s.Store.Store(userID, data); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
