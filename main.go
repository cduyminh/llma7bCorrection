package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	InitConfig(".")
	router := gin.Default()

	ctx := context.Background()
	redisClient := GetRedisClient()
	filePath := "document.docx"
	doc, err := ReadDocxFile(filePath)
	if err != nil {
		panic(err)
	}
	content := doc.Editable().rawContent

	CorrectingParagraph(ctx, content, redisClient)
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
