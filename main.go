package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	InitConfig(".")
	router := gin.Default()

	ctx := context.Background()
	redisClient := GetRedisClient()
	filePath := "document.docx"

	// Measure the time it takes to execute the read operation
	startTime := time.Now()

	_, err := readAndStoreInRedis(ctx, redisClient, filePath)
	if err != nil {
		panic(err)
	}

	elapsedTime := time.Since(startTime)
	if err != nil {
		panic(err)
	}

	doc, err := ReadDocxFile(filePath)
	if err != nil {
		panic(err)
	}

	content := doc.Editable().rawContent

	serializeEachParagraph(ctx, content, redisClient)

	// log.Println(doc.Editable().content)
	// editable := doc.Editable()

	// editable.WriteToFile("result.docx")

	fmt.Printf("Total time taken: %s\n", elapsedTime)

	router.POST("/upload", func(c *gin.Context) {
		// Parse the multipart form
		err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max memory
		if err != nil {
			c.String(http.StatusBadRequest, "Parse form err: %s", err.Error())
			return
		}

		// Retrieve the file from the form data
		file, header, err := c.Request.FormFile("upload[]")
		if err != nil {
			c.String(http.StatusBadRequest, "Retrieve file err: %s", err.Error())
			return
		}
		defer file.Close() // Always close the file after processing

		c.String(http.StatusOK, "Processed file: %s", header.Filename)
	})

	router.Run(":8080")
}
