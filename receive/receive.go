package main

// this doesn't work yet

import (
	"fmt"
	"log"
	"os"

	"github.com/pborman/getopt"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {

	rabbitMQUser := os.Getenv("RABBITMQ_DEFAULT_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_DEFAULT_PASS")
	rabbitMQHost := "127.0.0.1"

	messages := getopt.IntLong("number", 'n', 20, "Number of messages to remove from queue")
	queueName := getopt.StringLong("queue", 'q', "autoscaling", "Name of the rabbitMQ queue to add messages to")

	rabbitMQURI := fmt.Sprintf("amqp://%s:%s@%s:5672/", rabbitMQUser, rabbitMQPassword, rabbitMQHost)
	// rabbitMQURL := fmt.Sprintf("http://%s:15672/api/queues/%%2F/%s", rabbitMQHost, *queueName)

	help := getopt.Bool('h', "display help")
	getopt.SetParameters("")
	getopt.Parse()

	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	conn, err := amqp.Dial(rabbitMQURI)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		*queueName, // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	fmt.Printf("Messages on queue: %v", msgs)

	fmt.Printf("Taking %v messages off the queue", *messages)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
