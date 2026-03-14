package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	emailFrom := flag.String("emailfrom", "", "Email address of Ansible Server")
	emailTo := flag.String("emailto", "", "Email address of recipient for the Ansible report")
	flag.Parse()

	file, err := os.Open(*logFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	body, err := parseFailureReport(file)
	if err != nil {
		log.Fatal(err)
	}

	if err := sendEmail(body, *mailServer, *emailFrom, *emailTo); err != nil {
		log.Fatal(err)
	}
}

func parseFailureReport(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	var body strings.Builder
	for line := range bytes.Lines(data) {
		if !bytes.Contains(bytes.ToLower(line), []byte("failed")) {
			continue
		}

		hostPrefix, entryJSON, found := bytes.Cut(line, []byte("=>"))
		if !found {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(entryJSON, &entry); err != nil {
			continue
		}

		host, found := extractHost(hostPrefix)
		if !found {
			continue
		}

		_, _ = fmt.Fprintf(&body, "Host: %s\nMessage: %s\n\n", bytes.TrimSpace(host), entry.Msg)
	}

	return body.String(), nil
}

func extractHost(hostPrefix []byte) ([]byte, bool) {
	if _, afterOpen, found := bytes.Cut(hostPrefix, []byte("[")); found {
		if host, _, found := bytes.Cut(afterOpen, []byte("]")); found {
			return host, true
		}
	}

	host, _, found := bytes.Cut(hostPrefix, []byte("|"))
	if !found {
		return nil, false
	}

	return bytes.TrimSpace(host), true
}

func sendEmail(body string, mailserver string, fromemail string, toemail string) error {
	from := fromemail
	to := toemail

	msg := buildMessage(from, to, body)

	err := smtp.SendMail(mailserver,
		nil,
		from, []string{to}, msg)

	if err != nil {
		return fmt.Errorf("smtp send mail: %w", err)
	}

	log.Print("Email sent")
	return nil
}

func buildMessage(from string, to string, body string) []byte {
	return []byte("From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Ansible Log Report\n\n" +
		body)
}
