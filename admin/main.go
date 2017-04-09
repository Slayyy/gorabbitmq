package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/Slayyy/gorabbitmq/common"
	"github.com/streadway/amqp"
)

func readCommand(reader *bufio.Reader) (string, bool) {
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", false
	}

	text = strings.TrimSpace(text)
	return text, true

}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	common.PanicOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.PanicOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = common.DeclareAdminReceiverExchange(ch)
	common.PanicOnError(err, "Failed to declare a exchange")

	const adminQueueName = "admin_queue"
	q, err := ch.QueueDeclare(
		adminQueueName, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	common.PanicOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		common.AdminReceiverExchange, // exchange
		false,
		nil)

	common.PanicOnError(err, "Failed to bind a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	common.PanicOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msgs {
			log.Println(string(d.Body))
		}
	}()

	log.Println("----------ADMIN----------")
	reader := bufio.NewReader(os.Stdin)
	for {

		message, ok := readCommand(reader)
		if !ok {
			log.Println("Bad command")
			continue
		}

		err = ch.Publish(
			common.AdminInfoExchange, // exchange
			"",    // routing key
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,

				ContentType: "text/plain",
				Body:        []byte(message),
			})
		common.PanicOnError(err, "Failed to publish a message")
	}
}
