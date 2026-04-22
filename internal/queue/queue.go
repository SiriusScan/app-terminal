package queue

import (
	"fmt"
	"log/slog"

	"github.com/streadway/amqp"
)

func Send(queue string, message string) error {
	slog.Debug("sending message to queue", "queue", queue)
	conn, err := amqp.Dial("amqp://guest:guest@sirius-rabbitmq:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Declare queue with consistent settings
	_, err = ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	slog.Debug("message sent successfully", "queue", queue)
	return nil
}

func Listen(queue string, handler func(string)) {
	conn, err := amqp.Dial("amqp://guest:guest@sirius-rabbitmq:5672/")
	if err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to open channel", "error", err)
		return
	}
	defer ch.Close()

	// Declare queue with consistent settings
	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		slog.Error("failed to declare queue", "queue", queue, "error", err)
		return
	}

	slog.Info("listening on queue", "queue", queue)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		slog.Error("failed to register consumer", "queue", queue, "error", err)
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			slog.Debug("received message", "queue", queue, "size", len(d.Body))
			handler(string(d.Body))
		}
	}()

	slog.Info("queue consumer started, waiting for messages", "queue", queue)
	<-forever
}
