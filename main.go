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

	router.GET("/document", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello world!",
		})
	})

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
			if err := c.SaveUploadedFile(file, file.Filename); err != nil {
				c.String(http.StatusBadRequest, "Save uploaded file err: %s", err.Error())
				return
			}
		}

		c.String(http.StatusOK, "Uploaded successfully %d files.", len(files))
	})

	router.Run(":8080")
}
