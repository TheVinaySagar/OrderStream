package main

import (
	"awesomeProject/config"
	"awesomeProject/services/payment_service/internal"
)

func main() {
	kafkaClient := config.KafkaConsumer()
	p := config.KafkaProducer()

	h := internal.NewPaymentHandler()
	consumer := internal.NewPaymentService(kafkaClient, h, p)

	consumer.Start("order_created")
}
