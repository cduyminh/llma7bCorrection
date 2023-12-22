package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	producer    sarama.SyncProducer
	ctx         context.Context
	err         error
)

func main() {
	InitConfig(".")
	router := gin.Default()
	ctx = context.Background()
	redisClient = GetRedisClient()

	// Initialize Kafka producer
	producer, err = initKafkaProducer()
	if err != nil {
		log.Fatalf("Error initializing Kafka producer: %v", err)
	}

	// Create a Kafka consumer
	consumer, err := initKafkaConsumer()
	if err != nil {
		log.Fatalf("Error initializing Kafka consumer: %v", err)
	}
	defer consumer.Close()

	router.POST("/upload", func(c *gin.Context) {
		// Multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "Get form err: %s", err.Error())
			return
		}

		files := form.File["upload[]"]

		for _, file := range files {

			// You can now save the file or process it as needed
			if err := c.SaveUploadedFile(file, "Uploaded_"+file.Filename); err != nil {
				c.String(http.StatusBadRequest, "Save uploaded file err: %s", err.Error())

				return
			}

			filePath := "Uploaded_" + file.Filename
			doc, err := ReadDocxFile(filePath)
			if err != nil {
				panic(err)
			}
			content := doc.Editable().rawContent
			// CorrectingParagraph(ctx, content, redisClient)
			// Send the content to Kafka for asynchronous processing
			go sendContentToKafka(ctx, content)
			// Consume messages from the Kafka topic
			consumeMessages(consumer)
		}

		c.String(http.StatusOK, "Uploaded successfully %d files.", len(files))
	})

	router.Run(":8080")
}

func initKafkaConsumer() (sarama.Consumer, error) {
	config := sarama.NewConfig()
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		return nil, fmt.Errorf("error initializing Kafka consumer: %v", err)
	}
	return consumer, nil
}

func consumeMessages(consumer sarama.Consumer) {
	// Change the topic name to match the one you used for producing messages
	topic := "content-topic"

	// Create a partition consumer for the topic
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error creating partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	log.Printf("Consumer started for topic: %s", topic)

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			// Process the received message
			go processKafkaMessage(msg)
		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming message: %v", err)
		}
	}
}

func processKafkaMessage(msg *sarama.ConsumerMessage) {
	content := string(msg.Value)

	// Call the CorrectingParagraph function to process the content
	err := CorrectingParagraph(ctx, content, redisClient)
	if err != nil {
		log.Printf("Error processing message: %v", err)
	}
}

func initKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	brokers := []string{"localhost:9092"} // Update with your Kafka brokers
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("error initializing Kafka producer: %v", err)
	}

	return producer, nil
}

func sendContentToKafka(ctx context.Context, content string) {
	// Send the content to Kafka for asynchronous processing
	message := &sarama.ProducerMessage{
		Topic: "content-topic", // Change to your Kafka topic
		Value: sarama.StringEncoder(content),
	}

	_, _, err := producer.SendMessage(message)
	if err != nil {
		log.Printf("Error sending message to Kafka: %v", err)
	}
}
