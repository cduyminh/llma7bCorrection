package main

import (
	"context"
)

func main() {
	InitConfig(".")
	ctx := context.Background()
	redisClient := GetRedisClient()
	filePath := "document.docx"
	doc, err := ReadDocxFile(filePath)
	if err != nil {
		panic(err)
	}
	content := doc.Editable().rawContent

	CorrectingParagraph(ctx, content, redisClient)
}
