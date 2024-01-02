package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/joho/godotenv"
)

func SendEmail(to []string, fileName string, fileDataBuffer *bytes.Buffer) {
	// Email configuration
	from := "lamagong017@gmail.com"
	// Get the password from the environment variable
	env, _ := godotenv.Read("app.development.env")
	password := env["EMAIL_PASSWORD"]
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Encode the file data to base64
	encoded := base64.StdEncoding.EncodeToString(fileDataBuffer.Bytes())

	// Message details.
	var email bytes.Buffer

	// Headers
	email.WriteString(fmt.Sprintf("From: %s\r\n", from))
	email.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ";")))
	email.WriteString(fmt.Sprintf("Subject: %s\r\n", "Your document has been processed"))
	email.WriteString("MIME-Version: 1.0\r\n")
	email.WriteString("Content-Type: multipart/mixed; boundary=boundary\r\n\r\n")

	email.WriteString("--boundary\r\n")
	email.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	email.WriteString("Your document has been processed, please check with the following attached file" + "\r\n")

	email.WriteString("--boundary\r\n")
	email.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n", fileName))
	email.WriteString("Content-Type: application/octet-stream\r\n")
	email.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	email.WriteString(encoded + "\r\n")
	email.WriteString("--boundary--")

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)
	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, email.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}
