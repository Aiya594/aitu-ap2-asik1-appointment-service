package natspub

import (
	"log"

	"github.com/nats-io/nats.go"
)

type EventPublisher interface {
	Publish(subject string, payload []byte) error
	Close()
}

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Publisher{conn: nc}, nil
}

func (p *Publisher) Publish(subject string, payload []byte) error {
	err := p.conn.Publish(subject, payload)
	if err != nil {
		log.Printf("NATS publish error: subject=%s err=%v", subject, err)
		return err
	}

	return nil
}

func (p *Publisher) Close() {
	p.conn.Close()
}
