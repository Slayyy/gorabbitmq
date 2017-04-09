package common

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const (
	DoctorExchangeName = "doctor_exchange"
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
