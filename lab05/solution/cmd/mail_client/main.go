package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/smtp"
)

func main() {
	from := flag.String("from", "", "sender email")
	to := flag.String("to", "", "recipient email")
	format := flag.String("format", "txt", "message format: txt or html")
	subject := flag.String("subject", "Test message", "subject")
	body := flag.String("body", "Hello!", "message body")
	server := flag.String("server", "smtp.mail.ru:587", "smtp server")
	pass := flag.String("pass", "", "smtp password")
	flag.Parse()

	if *from == "" || *to == "" {
		log.Fatal("usage: mail_client -from <addr> -to <addr> -pass <pass> [-format txt|html] [-subject ...] [-body ...]")
	}

	contentType := "text/plain"
	if *format == "html" {
		contentType = "text/html"
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: %s; charset=UTF-8\r\n\r\n%s",
		*from, *to, *subject, contentType, *body,
	)

	host, _, _ := net.SplitHostPort(*server)

	c, err := smtp.Dial(*server)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}

	if err = c.StartTLS(&tls.Config{ServerName: host}); err != nil {
		log.Fatalf("starttls: %v", err)
	}

	if *pass != "" {
		if err = c.Auth(smtp.PlainAuth("", *from, *pass, host)); err != nil {
			log.Fatalf("auth: %v", err)
		}
	}

	if err = c.Mail(*from); err != nil {
		log.Fatalf("mail: %v", err)
	}
	if err = c.Rcpt(*to); err != nil {
		log.Fatalf("rcpt: %v", err)
	}

	w, err := c.Data()
	if err != nil {
		log.Fatalf("data: %v", err)
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		log.Fatalf("write: %v", err)
	}
	if err = w.Close(); err != nil {
		log.Fatalf("close data: %v", err)
	}

	c.Quit()
	fmt.Println("sent")
}
