package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type DeliveryPayload struct {
	OrderID uint   `json:"order_id"`
	UserID  uint   `json:"user_id"`
	Address string `json:"address"`
}

func main() {
	conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
	ch, _ := conn.Channel()
	ch.ExchangeDeclare("delivery-ex", "fanout", true, false, false, false, nil)

	q, _ := ch.QueueDeclare("", false, false, true, false, nil)
	ch.QueueBind(q.Name, "", "delivery-ex", false, nil)

	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	log.Println("🚚 Delivery Service đang lắng nghe...")

	for msg := range msgs {
		var data DeliveryPayload
		json.Unmarshal(msg.Body, &data)
		log.Printf("📦 Giao đơn hàng #%d đến user %d tại %s", data.OrderID, data.UserID, data.Address)
	}
}
