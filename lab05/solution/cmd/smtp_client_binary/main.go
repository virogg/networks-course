package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	from := flag.String("from", "", "sender email")
	to := flag.String("to", "", "recipient email")
	subject := flag.String("subject", "Test", "subject")
	body := flag.String("body", "Hello!", "body")
	image := flag.String("image", "", "path to image file")
	pass := flag.String("pass", "", "smtp password")
	server := flag.String("server", "smtp.mail.ru:587", "smtp server")
	flag.Parse()

	if *from == "" || *to == "" || *image == "" || *pass == "" {
		log.Fatal("usage: smtp_client_binary -from <addr> -to <addr> -pass <pass> -image <path> [-subject ...] [-body ...]")
	}

	imgData, err := os.ReadFile(*image)
	if err != nil {
		log.Fatalf("read image: %v", err)
	}

	const boundary = "----=_Part_boundary_42"
	imgName := filepath.Base(*image)

	// base64 in 76-char lines per RFC 2045
	raw := base64.StdEncoding.EncodeToString(imgData)
	var b64 strings.Builder
	for i := 0; i < len(raw); i += 76 {
		end := i + 76
		if end > len(raw) {
			end = len(raw)
		}
		b64.WriteString(raw[i:end])
		b64.WriteString("\r\n")
	}

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n",
		*from, *to, *subject, boundary,
	)
	textPart := fmt.Sprintf(
		"--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\n",
		boundary, *body,
	)
	imgPart := fmt.Sprintf(
		"--%s\r\nContent-Type: image/jpeg\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n",
		boundary, imgName, b64.String(),
	)
	closing := fmt.Sprintf("--%s--\r\n", boundary)

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
		fmt.Fprint(conn, s) //nolint:checkerr
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
		fmt.Fprint(conn, s) //nolint:checkerr
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

	fmt.Fprint(conn, headers+textPart+imgPart+closing) //nolint:checkerr
	fmt.Fprint(conn, "\r\n.\r\n")                      //nolint:checkerr
	expect("250")
	send("QUIT\r\n")
	expect("221")
	fmt.Println("done")
}
