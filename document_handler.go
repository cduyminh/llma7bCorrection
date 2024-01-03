package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func UploadFile(c *gin.Context) {
	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Get form err: %s", err.Error())
		return
	}

	files := form.File["upload[]"]
	emails, ok := form.Value["emails"]
	if !ok || len(emails) == 0 {
		c.String(http.StatusBadRequest, "Email addresses are required")
		return
	}

	for index, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.String(http.StatusBadRequest, "File open err: %s", err.Error())
			return
		}
		defer file.Close()

		// Read file data into a byte slice
		fileData, err := io.ReadAll(file)
		if err != nil {
			c.String(http.StatusBadRequest, "Read file err: %s", err.Error())
			return
		}

		// Send the file data to Kafka
		err = SendContentToKafka(ctx, fileHeader.Filename, emails, fileData)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error sending to Kafka: %s", err.Error())
			return
		}

		fmt.Println("files found: ", index+1)
		UpdateStatusToRedis(emails, fileHeader.Filename, 10)
	}

	c.String(http.StatusOK, "Processed successfully %d files.", len(files))
}

func StatusUpdate(c *gin.Context) {
	// Extract emails and filename from the query parameters
	emailQuery := c.Query("emails")
	filename := c.Query("filename")

	// Check if the parameters are provided
	if emailQuery == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and filename parameters are required"})
		fmt.Println(emailQuery)
		return
	}

	// Split the emailQuery into a slice of emails
	emails := strings.Split(emailQuery, ",")

	// Construct the Redis key
	emailKeyPart := strings.Join(emails, ":")
	docxKey := "status:" + emailKeyPart + ":" + filename

	// Retrieve the status from Redis
	statusValue, err := redisClient.Get(ctx, docxKey).Int()
	if err != nil {
		fmt.Println(emailQuery)
		// If the key does not exist or another error occurs, handle it
		if err == redis.Nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error retrieving status from Redis: %v", err)})
		}
		return
	}

	// Return the status as an integer
	c.JSON(http.StatusOK, gin.H{"status": statusValue})
}
