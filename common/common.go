package common

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const (
	DoctorExchangeName    = "doctor_exchange"
	AdminInfoExchange     = "admin_info_exchange"
	AdminReceiverExchange = "admin_receiver_exchange"
)

func PanicOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func DeclareDoctorRqQueue(name string, ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

}

func DeclareDoctorExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		DoctorExchangeName, // name
		"direct",           // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)

}

func DeclareAdminInfoExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		AdminInfoExchange, // name
		"fanout",          // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)

}

func DeclareAdminReceiverExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		AdminReceiverExchange, // name
		"fanout",              // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
}

func ListenForAdminInfo(ch *amqp.Channel) {
	err := DeclareAdminInfoExchange(ch)
	PanicOnError(err, "Failed to declare a exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	PanicOnError(err, "Failed to declare a queue")
	err = ch.QueueBind(
		q.Name,            // queue name
		"",                // routing key
		AdminInfoExchange, // exchange
		false,
		nil)
	PanicOnError(err, "Failed to bind a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	PanicOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msgs {
			log.Printf("Admin info: %s", d.Body)
			d.Ack(false)
		}
	}()

}
