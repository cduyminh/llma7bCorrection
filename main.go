package main

import (
	"context"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         context.Context
	op          *OrderPlacer
)

type OrderPlacer struct {
	producer         *kafka.Producer
	topic            string
	delivery_channel chan kafka.Event
}

func main() {
	InitConfig(".")
	router := gin.Default()
	ctx = context.Background()
	redisClient = GetRedisClient()

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "foo",
		"acks":              "all",
	})
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
	}

	// InitKafkaConsumer()
	op = NewOrderPlacer(p, Cfg.KakaTopic)

	go func() {
		InitKafkaConsumer()
	}()

	// Preview()
	router.POST("/upload", UploadFile)

	router.Run(":8080")
}

func Preview() {
	docx1, err := ReadDocxFile("Uploaded_document.docx")
	if err != nil {
		return
	}
	content1 := docx1.Editable().content
	log.Println(content1)
}
