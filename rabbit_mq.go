package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type AMQPChannelInterface interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

var rabbitConn *amqp.Connection
var rabbitChannel AMQPChannelInterface
var osExit = os.Exit

func InitRabbitMQ(cfg Config) {
	var err error
	rabbitConn, err = amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal(err)
	}

	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
}

func PublishToQueue(order PurchaseOrder) error {
	q, err := rabbitChannel.QueueDeclare("purchase_orders", true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(order)
	return rabbitChannel.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func ConsumeQueue(cfg Config) {
	q, err := rabbitChannel.QueueDeclare("purchase_orders", true, false, false, false, nil)
	if err != nil {
		log.Print(err)
		osExit(1)
		return
	}

	msgs, err := rabbitChannel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Print(err)
		osExit(1)
		return
	}

	for msg := range msgs {
		var po PurchaseOrder
		if err := json.Unmarshal(msg.Body, &po); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		err := SyncToDynamics(cfg, po)
		if err != nil {
			log.Printf("Sync failed: %v", err)
			ReportErrorToGlitchTip(cfg, po.ID, err)
		}
	}
}
