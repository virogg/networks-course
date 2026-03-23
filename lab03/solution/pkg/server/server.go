package server

import (
	"bufio"
	_ "embed"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var rootDir = "assets"

func Listen(port string) (net.Listener, func() error, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on port: %s: %v", port, err)
		return nil, nil, err
	}
	log.Printf("listening on port %s", port)
	return lis, lis.Close, nil
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	req, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("failed to read req: %v", err)
		return
	}

	req = strings.TrimSpace(req)
	log.Printf("Request: %v", req)

	parts := strings.Fields(req)
	path := parts[1]

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("failed to read req: %v", err)
			}
			break
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	path = strings.TrimPrefix(path, "/")
	path = filepath.Clean(path)

	fullPath := filepath.Join(rootDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		sendError(conn, 404, "Not Found")
		return
	}

	sendResponse(conn, getContentType(path), data)
}
