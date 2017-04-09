package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Slayyy/gorabbitmq/common"
	"github.com/streadway/amqp"
)

func readCommand(reader *bufio.Reader) (string, string, bool) {
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", "", false
	}

	text = strings.TrimSpace(text)
	s := strings.Split(text, " ")
	if len(s) < 2 {
		return "", "", false
	}
	return s[0], s[1], true

}

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	common.PanicOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.PanicOnError(err, "Failed to open a channel")
	defer ch.Close()
	common.ListenForAdminInfo(ch)

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	common.PanicOnError(err, "Failed to declare a queue")

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

	corrID := randomString(32)

	go func() {
		for d := range msgs {
			if corrID == d.CorrelationId {
				log.Printf("Received examination: %s", string(d.Body))
			}
		}
	}()

	err = common.DeclareDoctorExchange(ch)
	common.PanicOnError(err, "Failed to declare a exchange")

	log.Println("----------Doctor----------")
	reader := bufio.NewReader(os.Stdin)
	for {

		examinationType, surname, ok := readCommand(reader)
		if !ok {
			log.Println("Bad command")
			continue
		}

		err = ch.Publish(
			common.DoctorExchangeName, // exchange
			examinationType,           // routing key
			false,                     // mandatory
			false,                     // immediate
			amqp.Publishing{
				DeliveryMode:  amqp.Persistent,
				CorrelationId: corrID,
				ReplyTo:       q.Name,

				ContentType: "text/plain",
				Body:        []byte(surname),
			})
		common.PanicOnError(err, "Failed to publish a message")
	}

}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}
