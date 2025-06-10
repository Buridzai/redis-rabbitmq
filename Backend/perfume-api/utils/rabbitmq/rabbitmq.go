package rabbitmq

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var ch *amqp.Channel

func InitRabbitMQ() {
	var err error
	conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal("❌ Không thể kết nối RabbitMQ:", err)
	}

	ch, err = conn.Channel()
	if err != nil {
		log.Fatal("❌ Không thể tạo channel:", err)
	}

	ch.ExchangeDeclare("delivery-ex", "fanout", true, false, false, false, nil)
}

func Publish(queue string, payload interface{}) {
	body, _ := json.Marshal(payload)
	err := ch.Publish("delivery-ex", "", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Println("❌ Gửi message lỗi:", err) // 👈 THÊM LOG CHI TIẾT
		panic(err)                             // hoặc return error tùy ý
	}
}
