package main

import (
	"bytes"
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
		// PreviewRedis("samplefixgrammar (2).docx")
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

func PreviewRedis(filename string) error {
	fmt.Println("Start Reading")
	// Construct the Redis key using the filename
	docxKey := "corrected_docx:" + filename

	// Retrieve the DOCX file data from Redis
	data, err := redisClient.Get(ctx, docxKey).Bytes()
	if err != nil {
		fmt.Println("Error Reading: ", err)
		return fmt.Errorf("error retrieving DOCX from Redis: %v", err)
	}

	fmt.Println("Reading 1")

	// Check if data is retrieved
	if len(data) == 0 {
		fmt.Println("No data found for the given filename.")
		return nil
	}

	// Use ReadDocxFromMemory to read the DOCX file
	reader := bytes.NewReader(data)
	fmt.Println("Reading 2")
	doc, err := ReadDocxFromMemory(reader, int64(len(data)))
	if err != nil {
		return fmt.Errorf("error reading DOCX from memory: %v", err)
	}
	fmt.Println("Reading 3")
	// Print the content of the DOCX file
	fmt.Println("Content of the DOCX file:", doc.Editable().rawContent)

	return nil
}
