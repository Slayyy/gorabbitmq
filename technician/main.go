package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Slayyy/gorabbitmq/common"
	"github.com/streadway/amqp"
)

func parseCommandLine() ([]string, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("Bad args")
	}

	return os.Args[1:], nil

}
func main() {
	types, err := parseCommandLine()
	if err != nil {
		common.PanicOnError(err, "")
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	common.PanicOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.PanicOnError(err, "Failed to open a channel")
	defer ch.Close()

	common.ListenForAdminInfo(ch)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	common.PanicOnError(err, "Failed to set QoS")

	err = common.DeclareDoctorExchange(ch)
	common.PanicOnError(err, "Failed to declare a exchange")

	for _, s := range types {
		q, err := common.DeclareDoctorRqQueue(s, ch)
		common.PanicOnError(err, "Failed to declare a queue")

		log.Printf("Binding queue %s to exchange %s with routing key %s",
			q.Name, common.DoctorExchangeName, s)
		err = ch.QueueBind(
			q.Name, // queue name
			s,      // routing key
			common.DoctorExchangeName, // exchange
			false,
			nil)
		common.PanicOnError(err, "Failed to bind a queue")
		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		common.PanicOnError(err, "Failed to register a consumer")
		t := s
		go func() {
			for d := range msgs {
				log.Printf("%s: %s", t, d.Body)

				response := fmt.Sprintf("Examination: %s, Surname %s",
					t, d.Body)

				err = ch.Publish(
					"",        // exchange
					d.ReplyTo, // routing key
					false,     // mandatory
					false,     // immediate
					amqp.Publishing{
						ContentType:   "text/plain",
						CorrelationId: d.CorrelationId,
						Body:          []byte(response),
					})
				common.PanicOnError(err, "Failed to publish a message")
				d.Ack(false)

			}
		}()
	}

	log.Println("")
	log.Println("")
	log.Println("----------TECHNICIAN----------")

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	forever := make(chan bool)
	<-forever
}
