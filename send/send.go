package main

// App to add messages to the autoscaling queue, creating it if it does not already exist
// run this locally as follows
// kubectl port-forward sts/nfsaas-rabbitmq 5672
// kubectl port-forward sts/nfsaas-rabbitmq 15672
// go run send.go

// Prometheus metrics that will reflect the number of messages in this queue is rabbitmq_queue_messages{queue="autoscaling"}

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pborman/getopt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	Name     string `json:"name"`
	Messages int    `json:"messages"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// send a message to a particular rabbitMQ queue
func sendMessageToQueue(ch *amqp.Channel, body, queueName string) error {

	err := ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	return nil

}

// return the number of messages currently sitting in the queue
func getNumberOfMessagesOnQueue(url, user, password string) (int, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	var q Queue
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
		return 0, err
	}

	return q.Messages, nil

}

func main() {

	rabbitMQUser := os.Getenv("RABBITMQ_DEFAULT_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_DEFAULT_PASS")
	rabbitMQHost := "127.0.0.1"

	messages := getopt.IntLong("number", 'n', 20, "Number of messages to add to queue")
	queueName := getopt.StringLong("queue", 'q', "autoscaling", "Name of the rabbitMQ queue to add messages to")

	rabbitMQURI := fmt.Sprintf("amqp://%s:%s@%s:5672/", rabbitMQUser, rabbitMQPassword, rabbitMQHost)
	rabbitMQURL := fmt.Sprintf("http://%s:15672/api/queues/%%2F/%s", rabbitMQHost, *queueName)

	help := getopt.Bool('h', "display help")
	getopt.SetParameters("")
	getopt.Parse()

	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	fmt.Println(rabbitMQURI)
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

	for i := 1; i <= *messages; i++ {

		body := "Hello World!"

		if err = sendMessageToQueue(ch, body, q.Name); err != nil {
			failOnError(err, "Failed to publish a message")
		}

		log.Printf("Sent message %v \n", i)

	}

	// give it some time to register the new messages on the queue with the API
	time.Sleep(5 * time.Second)

	totalMessages, err := getNumberOfMessagesOnQueue(rabbitMQURL, rabbitMQUser, rabbitMQPassword)
	if err != nil {
		failOnError(err, "Failed to get total messages on queue")
	}

	fmt.Printf("Number of messages on queue is now %v\n", totalMessages)
}
