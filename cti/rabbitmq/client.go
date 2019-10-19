package rabbitmq

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

var (
	conn *amqp.Connection
	//TODO implement pool
)

func init() {

	user := os.Getenv("RABBITMQ_USER")
	if user == "" {
		user = "guest"
	}

	password := os.Getenv("RABBITMQ_PASS")
	if password == "" {
		password = "guest"
	}

	address := os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR")
	if address == "" {
		address = "rabbitmq"
	}

	tcp := os.Getenv("RABBITMQ_PORT_5672_TCP_PORT")
	if tcp == "" {
		tcp = "5672"
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, address, tcp)
	log.Printf("Connecting to RabbitMQ: %s...\n", url)

	max := 5
	var err error
	connectTicker := time.Tick(time.Second * 2)

LOOP:
	for {
		select {
		case <-connectTicker:
			conn, err = amqp.Dial(url)
			if err == nil {
				break LOOP
			}
			if max == 0 {
				log.Fatalln("Failed to connect to RabbitMQ")
			}
			max--
		}
	}
	failOnError(err, "Failed to connect to RabbitMQ")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// Send messages to the queue
func Send(queue string, message *bytes.Buffer) {

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		queue,    // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = ch.Publish(
		queue, // exchange
		"",    // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message.Bytes(),
		})

	failOnError(err, "Failed to publish a message")
}

// DeleteQueue - delete queue
func DeleteQueue(queue string) {
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	ch.ExchangeDelete(queue, false, false)
	failOnError(err, "Failed to delete a queue")
}

//Close the connection
func Close() {
	conn.Close()
}
