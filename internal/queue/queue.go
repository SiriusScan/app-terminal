package queue

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func Send(queue string, message string) error {
	log.Printf("Sending message to queue %s: %s", queue, message)
	conn, err := amqp.Dial("amqp://guest:guest@sirius-rabbitmq:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
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
		return fmt.Errorf("failed to declare queue: %v", err)
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
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("Message sent successfully to queue %s", queue)
	return nil
}

func Listen(queue string, handler func(string)) {
	conn, err := amqp.Dial("amqp://guest:guest@sirius-rabbitmq:5672/")
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open channel: %v", err)
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
		log.Printf("Failed to declare queue: %v", err)
		return
	}

	log.Printf("Listening on queue: %s", queue)

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
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received message: %s", d.Body)
			handler(string(d.Body))
		}
	}()

	log.Printf("Queue consumer started. Waiting for messages...")
	<-forever
}
