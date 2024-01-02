package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// CorrectionResult represents the result of the correction API call.
type CorrectionResult struct {
	Original  string `json:"original"`
	Corrected string `json:"corrected"`
}

func SendContentToKafka(ctx context.Context, filename string, fileData []byte) error {
	// Assuming op is a properly initialized OrderPlacer instance
	return op.placeOrder(Cfg.KakaTopic, filename, fileData)
}

func CorrectingParagraph(docx *Docx, filename string) error {
	// Construct the Redis key using the filename
	docxKey := "corrected_docx:" + filename

	// Check if the key already exists in Redis
	exists, err := redisClient.Exists(ctx, docxKey).Result()
	if err != nil {
		return fmt.Errorf("error checking Redis for existing key: %v", err)
	}

	// If the key exists, return nil as no further action is needed
	if exists > 0 {
		fmt.Println("File Existed!")
		return nil
	}

	contents := ExtractContentBetweenWTags(docx.content)
	// // Call the correction API on the entire paragraph
	corrected, err := callCorrectionOpenAiApiOnParagraph(contents)
	if err != nil {
		return fmt.Errorf("error calling correction API: %v", err)
	}

	for _, content := range corrected {
		// key := "sentence:" + content.Original
		correction := content.Corrected
		if correction != "CORRECT" {
			// Split the original and corrected sentences into words
			originalWords := strings.Fields(content.Original)
			correctedWords := strings.Fields(content.Corrected)

			// Check if both sentences have at least one word each
			if len(originalWords) > 0 && len(correctedWords) > 0 {
				// Check if the lowercase of the first word of the original matches the lowercase of the first word of the corrected
				if strings.EqualFold(strings.ToLower(originalWords[0]), strings.ToLower(correctedWords[0])) {
					// Replace the first word of the corrected with the first word of the original
					correctedWords[0] = originalWords[0]
					// Recreate the corrected sentence with the updated first word
					correction = strings.Join(correctedWords, " ")
				}
			}

			// Replace the original text with the corrected text in the DOCX
			docx.Replace(content.Original, correction, -1)
		}

		// // Write the corrected paragraph to Redis
		// err = redisClient.Set(ctx, key, correction, 0).Err()
		// if err != nil {
		// 	return fmt.Errorf("error writing to Redis: %v", err)
		// }
	}

	// Edge cases
	docx.Replace("..", ".", -1)
	docx.Replace("..", "...", -1)
	docx.Replace(").).", ").", -1)
	// docx.WriteToFile("./new_result_1.docx")
	// Once all corrections are applied, write the DOCX to an in-memory buffer
	var buf bytes.Buffer
	err = docx.Write(&buf)
	if err != nil {
		return fmt.Errorf("error writing DOCX to buffer: %v", err)
	}

	// Convert the buffer to a byte slice
	docxBytes := buf.Bytes()

	// Save the byte slice to Redis
	err = redisClient.Set(ctx, docxKey, docxBytes, 0).Err()
	if err != nil {
		return fmt.Errorf("error writing DOCX to Redis: %v", err)
	}

	return nil
}

func callCorrectionOpenAiApiOnParagraph(paragraphs []string) ([]CorrectionResult, error) {
	// Initialize the OpenAI client with your API token
	client := openai.NewClient(Cfg.OpenApiKey)

	// Define the model to use (GPT-3.5 Turbo)
	model := openai.GPT3Dot5TurboInstruct

	// Define the maximum number of paragraphs to process in each batch
	batchSize := 20

	// Prepare messages for the chat completion
	var completions []CorrectionResult

	for i := 0; i < len(paragraphs); i += batchSize {
		end := i + batchSize
		if end > len(paragraphs) {
			end = len(paragraphs)
		}
		batch := paragraphs[i:end]

		var messages []string
		for _, paragraph := range batch {
			messages = append(messages, "Help me correct the spelling errors in this sentence.\nNote: Fix only spelling errors of words and nothing else. Do not change word cases. Do not add or remove or replace any punctuations, brackets, commas, or full stops. Do not capitalize my first word if it is not already capitalized. If the sentence is correct then return 'CORRECT' else return the corrected sentence.\nExample 1:\nspecificalities of vigor in the context of IIs\nThe correct sentence is:\nspecificities of vigor in the context of IIs\nExample 2:\nJava, Open source software\nThe correct sentence is:\nCORRECT\nExample 3:\n"+paragraph+"\nThe correct sentence is:\n")
		}

		// Create a chat completion request for the batch
		req := openai.CompletionRequest{
			Model:       model,
			Prompt:      messages,
			Temperature: 0.3,
			MaxTokens:   2000,
		}

		// Make the chat completion API call for the batch
		resp, err := client.CreateCompletion(context.Background(), req)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		// Process the API response and extract corrected text for the batch
		for j, choice := range resp.Choices {
			correctedText := strings.TrimSpace(choice.Text)
			completions = append(completions, CorrectionResult{
				Original:  strings.TrimSpace(batch[j]),
				Corrected: correctedText,
			})
		}
	}

	// Log the results
	// for i, completion := range completions {
	// 	log.Printf("Result %d: Original: %s\n", i+1, completion.Original)
	// 	log.Printf("Result %d: Corrected: %s\n", i+1, completion.Corrected)
	// }

	return completions, nil
}

func ExtractContentBetweenWTags(inputXML string) []string {
	// Define a regular expression pattern to match the content between <w:t> and </w:t> tags.
	re := regexp.MustCompile(`<[^>]*>(.*?)</[^>]*>`)

	// Find all matches of the pattern in the input XML.
	matches := re.FindAllStringSubmatch(inputXML, -1)

	// Extract the content between <w:t> and </w:t> tags and store them in an array.
	var contentArray []string
	for _, match := range matches {
		xmlTagPattern := "<[^>]+>"
		content := match[1]

		if !regexp.MustCompile(xmlTagPattern).MatchString(content) {
			trimmedContent := adjustSentence(content)
			// Split the trimmed content into words
			words := strings.Fields(trimmedContent)
			if len(words) > 3 {
				contentArray = append(contentArray, trimmedContent)
			}
		}
	}

	return contentArray
}

// Helper function to adjust a sentence to start and end with a character from "a-z, A-Z, 0-9"
func adjustSentence(s string) string {
	// Define a regular expression pattern to remove unwanted characters from the beginning and end
	// This pattern matches any characters that are not in the range a-z, A-Z, 0-9 at the beginning or end of the string
	re := regexp.MustCompile(`(^[^a-zA-Z0-9]+)|([^a-zA-Z0-9]+$)`)

	// Replace unwanted characters at the beginning and end with a space
	adjusted := re.ReplaceAllString(s, " ")

	// Trim any remaining leading and trailing spaces
	return strings.TrimSpace(adjusted)
}

// func callCorrectionApiOnParagraph(paragraphs []string) (result []string, err error) {
// 	// API endpoint URL
// 	apiURL := "https://trusting-inherently-feline.ngrok-free.app/generate_code"

// 	// Prepare prompts for the API call
// 	var prompts []string
// 	maxPrompts := 5 // Set the maximum number of prompts

// 	for i, paragraph := range paragraphs {
// 		prompts = append(prompts, "Help me correct the spelling errors in this sentence: "+paragraph)
// 		if i >= maxPrompts-1 {
// 			break
// 		}
// 	}

// 	// Combine prompts into URL parameters
// 	urlParams := strings.Join(prompts, "&prompts=")
// 	urlParams = strings.ReplaceAll(urlParams, " ", "%20")

// 	// Construct the final API request URL
// 	apiURL = fmt.Sprintf("%s?prompts=%s", apiURL, urlParams)

// 	// Make the HTTP request
// 	response, err := http.Get(apiURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer response.Body.Close()

// 	// Read the response body
// 	body, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Parse the result if needed (split by newline)
// 	result = strings.Split(string(body), "\n")

// 	log.Printf("URL: %s", apiURL)

// 	log.Printf("Result: %d lines", len(result))

// 	// Print all results
// 	for i, line := range result {
// 		fmt.Printf("Result %d: %s\n", i+1, line)
// 	}

// 	return result, nil
// }
