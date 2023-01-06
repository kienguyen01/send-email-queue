package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Message struct {
	SenderEmail   string
	SenderName    string
	ReceiverEmail string
	ReceiverName  string
	Body          string
	Subject       string
}

func main() {
	client, err := NewELKClient("localhost", "9200")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Consumer - Connecting to the Rabbit channel")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		log.Println("Error in connection")
	}

	//close connection at end of program
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	defer ch.Close()

	msgs, err := ch.Consume(
		"test-queue", // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)

	//process messages from this channel
	//forever will BLOCK the main.go program until it received a value from the channel
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var m Message
			json.Unmarshal(d.Body, &m)

			SendEmail(m)
			client.SendLog("Receiving", m)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func SendEmail(m Message) {
	log.Println("send email triggered")
	from := mail.NewEmail(m.SenderName, m.SenderEmail)
	subject := m.Subject
	to := mail.NewEmail(m.ReceiverName, m.ReceiverEmail)
	plainTextContent := m.Body
	htmlContent := ""
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
