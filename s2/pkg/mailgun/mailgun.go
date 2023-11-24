package mailgun

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

// Mailgun
// contains variables for connection to mailgun.
type Mailgun struct {
	APIKEY string
	Client *mailgun.MailgunImpl
	Sender string
}

// NewConnection
// opens a new connection for mailgun service.
func NewConnection(cfg Config) *Mailgun {
	return &Mailgun{
		APIKEY: cfg.APIKEY,
		Client: mailgun.NewMailgun(cfg.Domain, cfg.APIKEY),
		Sender: cfg.Sender,
	}
}

// Send
// emails with mailgun client.
func (m *Mailgun) Send(content, subject, receiver string) error {
	message := m.Client.NewMessage(m.Sender, subject, content, receiver)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, _, err := m.Client.Send(ctx, message)

	return err
}

type Config struct {
	Domain string `koanf:"domain"`
	APIKEY string `koanf:"api_key"`
	Sender string `koanf:"sender"`
}
