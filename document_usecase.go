package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func splitIntoSentences(paragraph string) []string {
	// Split paragraph into sentences
	var sentences []string

	// Split the paragraph into sentences based on common sentence-ending punctuation marks.
	sentenceDelimiters := []string{".", "!", "?"}
	paragraph = strings.ReplaceAll(paragraph, "\n", " ") // Remove line breaks
	paragraph = strings.TrimSpace(paragraph)
	sentenceParts := strings.FieldsFunc(paragraph, func(r rune) bool {
		return strings.ContainsRune(sentenceDelimiters[0]+sentenceDelimiters[1]+sentenceDelimiters[2], r)
	})

	if len(sentenceParts) == 0 {
		return []string{paragraph} // Return the whole paragraph if no sentence-ending punctuation is found.
	}

	currentSentence := sentenceParts[0]
	for i := 1; i < len(sentenceParts); i++ {
		currentSentence += " " + sentenceParts[i]
		if strings.ContainsAny(sentenceParts[i], ".!?") {
			sentences = append(sentences, strings.TrimSpace(currentSentence))
			currentSentence = ""
		}
	}

	if currentSentence != "" {
		sentences = append(sentences, strings.TrimSpace(currentSentence))
	}

	return sentences
}

func CorrectingParagraph(ctx context.Context, content string, redisClient *redis.Client) error {
	// Split content into paragraphs
	paragraphs := strings.Split(content, "\n")

	// Prepare slices for storing chunks and their corresponding original paragraphs
	var allChunks []string
	var originalParagraphs []string

	// Split each paragraph into chunks and store them
	for _, paragraph := range paragraphs {
		chunks := splitIntoSentences(paragraph)
		allChunks = append(allChunks, chunks...)
		originalParagraphs = append(originalParagraphs, paragraph)
	}

	log.Println(len(allChunks))

	// Call the correction API on all chunks at once
	englishCorrections, err := callCorrectionApiOnParagraph(allChunks)
	if err != nil {
		return fmt.Errorf("error calling correction API: %v", err)
	}

	// Iterate through the set and write each corrected paragraph to Redis
	for index, correction := range englishCorrections {
		key := "paragraph:" + originalParagraphs[index]
		value := correction

		err := redisClient.Set(ctx, key, value, 0).Err()
		if err != nil {
			return fmt.Errorf("error writing to Redis: %v", err)
		}
	}

	return nil
}

func callCorrectionApiOnParagraph(paragraphs []string) (result []string, err error) {
	url := "https://3c92-103-253-89-37.ngrok-free.app/generate_code"

	// Prepare prompts for the API call
	var prompts []string
	for i, paragraph := range paragraphs {
		prompts = append(prompts, "Help me correct english of this sentence remember to only answer with the corrected version of this sentence: "+paragraph+"")
		// prompts = append(prompts, "correct english of this sentence: "+"hello")
		if i > 5 {
			break
		}
	}

	// Combine prompts into URL parameters
	urlParams := strings.Join(prompts, "&prompts=")
	urlParams = strings.ReplaceAll(urlParams, " ", "%20")
	// Make the HTTP request
	apiURL := fmt.Sprintf("%s?prompts=%s", url, urlParams)
	log.Println("apiUrl:")
	log.Println(apiURL)
	response, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Parse the result if needed
	result = strings.Split(string(body), "\n")
	log.Println("Result:")
	log.Println(result)

	return result, nil
}

func readAndStoreInRedis(ctx context.Context, redisClient *redis.Client, filePath string) (replaceDocx *ReplaceDocx, err error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return
	}

	printFileStats(fileInfo)

	fileKey := generateFileKey(fileInfo)

	startReadTime := time.Now()

	fileData, err := redisClient.Get(ctx, fileKey).Result()
	if err == nil {
		fmt.Println("Reading data from Redis...")
		// replaceDocx.Editable().SetContent(fileData)

		fmt.Println("File Data:", fileData)
	} else {
		fmt.Println("Reading data from DOCX file...")

		replaceDocx, err = ReadDocxFile(filePath)
		if err != nil {
			return
		}

		content := replaceDocx.Editable().rawContent

		err = redisClient.Set(ctx, fileKey, content, 0).Err()
		if err != nil {
			return
		}
	}

	elapsedReadTime := time.Since(startReadTime)
	fmt.Printf("Time taken for read operation: %s\n", elapsedReadTime)

	return
}
