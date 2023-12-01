package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	InitConfig(".")

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

}
