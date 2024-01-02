package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Get form err: %s", err.Error())
		return
	}

	files := form.File["upload[]"]
	emails, ok := form.Value["email"]
	if !ok || len(emails) == 0 {
		c.String(http.StatusBadRequest, "Email addresses are required")
		return
	}

	for _, fileHeader := range files {
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
	}

	c.String(http.StatusOK, "Processed successfully %d files.", len(files))
}
