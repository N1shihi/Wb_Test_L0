package main

import (
	"Wb_Test_L0/internal/config"
	"Wb_Test_L0/internal/kafka"
	"Wb_Test_L0/internal/models"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func main() {

	cfg := config.MustLoad()
	log.Println("Loaded config:", cfg)
	n := cfg.Producer.NmbOfOrders
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	baseOrder := models.Order{
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       "2021-11-26T06:22:19Z",
		OofShard:          "1",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:     9934930,
				Price:      453,
				Name:       "Mascaras",
				Sale:       30,
				Size:       "0",
				TotalPrice: 317,
				NmID:       2389212,
				Brand:      "Vivienne Sabo",
				Status:     202,
			},
		},
	}

	for i := 0; i < n; i++ {
		order := baseOrder

		u := uuid.New().String()
		order.OrderUID = fmt.Sprintf("%s", u)
		order.TrackNumber = fmt.Sprintf("%s", u)
		order.Payment.Transaction = fmt.Sprintf("%s", u)

		for idx := range order.Items {
			order.Items[idx].Rid = fmt.Sprintf("%s-%s-%d", "rid", u, idx)
			order.Items[idx].TrackNumber = order.TrackNumber
		}

		if err := producer.Produce(order, cfg.Kafka.Topic); err != nil {
			log.Printf("failed to produce message %d: %v", i+1, err)
		} else {
			log.Printf("queued message %d order_uid=%s", i+1, order.OrderUID)
		}

		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("All messages queued, flushing and exiting")
}
