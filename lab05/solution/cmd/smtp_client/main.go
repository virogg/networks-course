package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	from := flag.String("from", "", "sender email")
	to := flag.String("to", "", "recipient email")
	subject := flag.String("subject", "Test", "subject")
	body := flag.String("body", "Hello!", "body")
	pass := flag.String("pass", "", "smtp password")
	server := flag.String("server", "smtp.mail.ru:587", "smtp server")
	flag.Parse()

	if *from == "" || *to == "" || *pass == "" {
		log.Fatal("usage: smtp_client -from <addr> -to <addr> -pass <pass> [-subject ...] [-body ...]")
	}

	host, _, _ := net.SplitHostPort(*server)

	conn, err := net.Dial("tcp", *server)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close() //nolint:checkerr

	r := bufio.NewReader(conn)

	expect := func(code string) {
		line, _ := r.ReadString('\n')
		fmt.Print("<- ", line)
		if !strings.HasPrefix(line, code) {
			log.Fatalf("expected %s, got: %s", code, line)
		}
	}
	send := func(s string) {
		fmt.Print("-> ", s)
		fmt.Fprint(conn, s)
	}
	readMultiline := func() {
		for {
			line, _ := r.ReadString('\n')
			fmt.Print("<- ", line)
			if len(line) >= 4 && line[3] == ' ' {
				break
			}
		}
	}

	expect("220")
	send("EHLO localhost\r\n")
	readMultiline()

	// STARTTLS
	send("STARTTLS\r\n")
	expect("220")
	tlsConn := tls.Client(conn, &tls.Config{ServerName: host})
	if err := tlsConn.Handshake(); err != nil {
		log.Fatalf("tls handshake: %v", err)
	}
	conn = tlsConn
	r = bufio.NewReader(conn)
	send = func(s string) {
		fmt.Print("-> ", s)
		fmt.Fprint(conn, s)
	}

	send("EHLO localhost\r\n")
	readMultiline()

	// AUTH LOGIN
	send("AUTH LOGIN\r\n")
	expect("334")
	send(base64.StdEncoding.EncodeToString([]byte(*from)) + "\r\n")
	expect("334")
	send(base64.StdEncoding.EncodeToString([]byte(*pass)) + "\r\n")
	expect("235")

	send(fmt.Sprintf("MAIL FROM:<%s>\r\n", *from))
	expect("250")
	send(fmt.Sprintf("RCPT TO:<%s>\r\n", *to))
	expect("250")
	send("DATA\r\n")
	expect("354")

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		*from, *to, *subject, *body,
	)
	fmt.Fprint(conn, msg)         //nolint:checkerr
	fmt.Fprint(conn, "\r\n.\r\n") //nolint:checkerr
	expect("250")
	send("QUIT\r\n")
	expect("221")
	fmt.Println("done")
}
