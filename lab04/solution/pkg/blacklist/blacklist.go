package blacklist

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/url"
	"os"
	"strings"
)

type Blacklist map[string]bool

func Load(path string) Blacklist {
	bl := make(map[string]bool)
	f, err := os.Open(path)
	if err != nil {
		log.Printf("load blacklist: %v", err)
		return bl
	}
	defer f.Close() //nolint:errcheck

	websites := struct {
		Addrs []string `json:"blacklist"`
	}{}

	if err := json.NewDecoder(f).Decode(&websites); err != nil {
		log.Printf("blacklist: %v", err)
	}
	for _, addr := range websites.Addrs {
		if _, err := url.Parse(addr); err != nil {
			log.Printf("blacklist: addr=%s is not valid: %v", addr, err)
			continue
		}
		bl[strings.ToLower(addr)] = true
	}

	return bl
}

func (bl Blacklist) IsBlocked(target string) bool {
	u, err := url.Parse(target)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if bl[strings.ToLower(target)] {
		return true
	}
	parts := strings.Split(host, ".")
	for i := range parts {
		if bl[strings.Join(parts[i:], ".")] {
			return true
		}
	}
	return false
}

//go:embed blacklist.tmpl
var blacklistTemplate string

func Respond(target string) (string, error) {
	tmpl, err := template.New("response").Parse(blacklistTemplate)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, struct {
		Address string
	}{target})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
