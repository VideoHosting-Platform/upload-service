package queue

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	User      string `env:"RABBITMQ_USER"`
	Password  string `env:"RABBITMQ_PASSWORD"`
	Host      string `env:"RABBITMQ_HOST"`
	Port      string `env:"RABBITMQ_PORT" env-default:"5672"`
	QueueName string `env:"RABBITMQ_QUEUE_NAME"`
}

type Queue struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string

	chMutex sync.Mutex
}

func New(cfg *Config) (*Queue, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Queue{
		conn:      conn,
		ch:        ch,
		queueName: cfg.QueueName,
	}, nil
}

func (q *Queue) Publish(ctx context.Context, body []byte) error {
	q.chMutex.Lock()
	defer q.chMutex.Unlock()

	err := q.ch.Publish(
		"", // exchange
		q.queueName,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	return err
}

func (q *Queue) Close() error {
	q.ch.Close()
	return q.conn.Close()
}
