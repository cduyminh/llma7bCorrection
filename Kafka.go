package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func NewOrderPlacer(p *kafka.Producer, topic string) *OrderPlacer {
	return &OrderPlacer{
		producer:         p,
		topic:            topic,
		delivery_channel: make(chan kafka.Event, 10000),
	}
}

func (op *OrderPlacer) placeOrder(orderType string, filename string, fileData []byte) error {
	// Create a JSON object with the filename and file data
	payload := map[string]interface{}{
		"filename": filename,
		"data":     fileData,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Produce the JSON string as the Kafka message value
	err = op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &op.topic, Partition: kafka.PartitionAny},
		Value:          payloadJSON,
	},
		op.delivery_channel,
	)
	if err != nil {
		log.Fatal(err)
	}
	<-op.delivery_channel
	return nil
}

func InitKafkaConsumer() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "foo",
		"auto.offset.reset": "smallest",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = consumer.Subscribe(Cfg.KakaTopic, nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		ev := consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			// Parse the JSON object to get filename and file data
			var payload struct {
				Filename string `json:"filename"`
				Data     []byte `json:"data"`
			}
			err := json.Unmarshal(e.Value, &payload)
			if err != nil {
				fmt.Printf("Error parsing message: %s\n", err)
				continue
			}

			// Convert the file data to an io.ReaderAt
			reader := bytes.NewReader(payload.Data)

			// Use ReadDocxFromMemory
			doc, err := ReadDocxFromMemory(reader, int64(len(payload.Data)))
			if err != nil {
				fmt.Printf("Error processing docx: %s\n", err)
				continue
			}

			// Process the doc as needed
			fmt.Println("KAFKA: Received and Processing docx - ", payload.Filename)

			CorrectingParagraph(doc.Editable(), payload.Filename)

		case *kafka.Error:
			fmt.Printf("%v\n", e)
		}
	}
}
