package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type LogEntry struct {
	Changed bool   `json:"changed"`
	Item    string `json:"item"`
	Msg     string `json:"msg"`
}

func main() {
	logFile := flag.String("logfile", "ansible.log", "Path to the Ansible log file")
	mailServer := flag.String("mailserver", "vmta.example.com:25", "Mail server to use for email relay")
	emailFrom := flag.String("emailfrom", "ansible.parser@example.com", "Email address of Ansible Server")
	emailTo := flag.String("emailto", "foo@example.com", "Email address of recipient for the Ansible report")
	flag.Parse()

	file, err := os.Open(*logFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "failed:") {
			parts := strings.SplitN(line, "=>", 2)
			if len(parts) != 2 {
				continue
			}
			var entry LogEntry
			err := json.Unmarshal([]byte(parts[1]), &entry)
			if err != nil {
				continue
			}
			hostParts := strings.SplitN(parts[0], "|", 2)
			buffer.WriteString(fmt.Sprintf("Host: %s\nMessage: %s\n\n", strings.TrimSpace(hostParts[1]), entry.Msg))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	sendEmail(buffer.String(), *mailServer, *emailFrom, *emailTo)
}

func sendEmail(body string, mailserver string, fromemail string, toemail string) {
	from := fromemail
	to := toemail

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Ansible Log Report\n\n" +
		body

	err := smtp.SendMail(mailserver,
		nil,
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("Email sent")
}
