package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ExtractTarget(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		path = "http://" + path
	}
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	return path
}

func Forward(r *http.Request, targetURL string, extra http.Header) (*http.Response, []byte, error) {
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	for k, vs := range r.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	for k, vs := range extra {
		for _, v := range vs {
			req.Header.Set(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}
	return resp, body, nil
}

func WriteResp(w http.ResponseWriter, resp *http.Response, body []byte) {
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body) //nolint:errcheck
}
