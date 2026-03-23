package server

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"path/filepath"
	"strings"
)

//go:embed response.tmpl
var responseTemplate string

type response struct {
	Code   int
	Reason string
}

func makeResponse(resp response) (string, error) {
	tmpl, err := template.New("response").Parse(responseTemplate)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, resp)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func sendError(conn net.Conn, code int, reason string) {
	body, err := makeResponse(response{
		Code:   code,
		Reason: reason,
	})
	if err != nil {
		log.Fatalf("failed to parse response: %v", err)
	}
	header := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: text/html\r\nContent-Length: %d\r\nConnection: close\r\n\r\n",
		code, reason, len(body))
	conn.Write([]byte(header))
	conn.Write([]byte(body))
}

func sendResponse(conn net.Conn, contentType string, body []byte) {
	header := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n",
		contentType, len(body))
	conn.Write([]byte(header))
	conn.Write(body)
}

var contentType = map[string]string{
	".html": "text/html; charset=utf-8",
	".htm":  "text/html; charset=utf-8",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".txt":  "text/plain; charset=utf-8",
	".pdf":  "application/pdf",
}

func getContentType(path string) string {
	if ct, ok := contentType[strings.ToLower(filepath.Ext(path))]; ok {
		return ct
	}
	return "application/octet-stream"
}
