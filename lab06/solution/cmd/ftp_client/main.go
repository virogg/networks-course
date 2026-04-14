package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	ftpclient "github.com/virogg/networks-course/lab06/solution/internal/transport/ftp/client"
	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

func main() {
	host := flag.String("host", "127.0.0.1", "FTP server host")
	port := flag.Int("port", 8081, "FTP server port")
	user := flag.String("user", "test", "FTP username")
	pass := flag.String("pass", "test", "FTP password")
	flag.Parse()

	addr := net.JoinHostPort(*host, strconv.FormatInt(int64(*port), 10))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	c := ftpclient.New(conn)
	defer c.Close()

	code, msg, _ := c.ReadResponse()
	fmt.Println(msg)
	if code != ftp.StatusServiceReady {
		log.Fatal("unexpected welcome")
	}

	code, msg, _ = c.Cmd("USER " + *user)
	fmt.Println(msg)
	code, msg, _ = c.Cmd("PASS " + *pass)
	fmt.Println(msg)
	if code != ftp.StatusLoginSuccessful {
		log.Fatal("login failed")
	}

	fmt.Println("\nCommands: ls, get <file>, put <file>, pwd, cd <dir>, quit")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("ftp> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "ls":
			data, err := c.List()
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			fmt.Print(string(data))
		case "get":
			if len(parts) < 2 {
				fmt.Println("usage: get <filename>")
				continue
			}
			data, err := c.Get(parts[1])
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			os.WriteFile(parts[1], data, 0644)
			fmt.Printf("saved %s (%d bytes)\n", parts[1], len(data))
		case "put":
			if len(parts) < 2 {
				fmt.Println("usage: put <filename>")
				continue
			}
			fileData, err := os.ReadFile(parts[1])
			if err != nil {
				fmt.Println("read file error:", err)
				continue
			}
			if err := c.Put(parts[1], fileData); err != nil {
				fmt.Println("error:", err)
				continue
			}
			fmt.Printf("uploaded %s (%d bytes)\n", parts[1], len(fileData))
		case "pwd":
			_, msg, _ := c.Cmd("PWD")
			fmt.Println(msg)
		case "cd":
			if len(parts) < 2 {
				fmt.Println("usage: cd <dir>")
				continue
			}
			_, msg, _ := c.Cmd("CWD " + parts[1])
			fmt.Println(msg)
		case "quit", "exit":
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
