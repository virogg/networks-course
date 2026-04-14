package main

import (
	"flag"
	"fmt"
	"log"

	ftpserver "github.com/virogg/networks-course/lab06/solution/internal/transport/ftp"
)

func main() {
	port := flag.Int("port", 8081, "FTP server port")
	rootDir := flag.String("root", "./ftpdata", "root directory for FTP files")
	user := flag.String("user", "test", "FTP username")
	pass := flag.String("pass", "test", "FTP password")
	flag.Parse()

	srv := ftpserver.NewServer(*rootDir, *user, *pass)
	log.Fatal(srv.ListenAndServe(fmt.Sprintf(":%d", *port)))
}
