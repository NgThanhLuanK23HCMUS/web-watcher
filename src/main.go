// go.mod: module tinywaf
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Regular expressions for detecting various attack patterns
var (
	sqlRe  = regexp.MustCompile(`(?i)(?:' *or *'?\d*'? *=|union[\s%20]+select|or[\s%20]+[0-9]+[\s%20]*=|select(?:[\s%20]+\*?[\s%20]*|\s*\*\s*)from|select[\s%20]+from|or[\s%20]+1\s*=\s*1|'[\s%20]*or[\s%20]*'|'\s*or\s*'[^']*'\s*=\s*'|--|\+.*=|\bxp_|\bdrop\b|;\s*$)`)
	xssRe  = regexp.MustCompile(`(?i)(<script|javascript:|onload=|onerror=|<img|alert\s*\(|prompt\s*\()`)
	pathRe = regexp.MustCompile(`(?i)(\.\.\/|\.\.\%2f|\/etc\/passwd|\/bin\/bash|cmd\.exe)`)
	cmdRe  = regexp.MustCompile(`(?i)(\|\s*cmd|\|\s*bash|\|\s*powershell|;&|;\s*cmd|\$\{IFS\}|system\s*\()`)
)

func checkMaliciousContent(input string) error {
	// Check for SQL Injection patterns
	if sqlRe.MatchString(input) {
		return fmt.Errorf("SQL injection pattern detected")
	}

	// Check for XSS patterns
	if xssRe.MatchString(input) {
		return fmt.Errorf("XSS pattern detected")
	}

	// Check for Path Traversal
	if pathRe.MatchString(input) {
		return fmt.Errorf("path traversal pattern detected")
	}

	// Check for Command Injection
	if cmdRe.MatchString(input) {
		return fmt.Errorf("command injection pattern detected")
	}

	return nil
}

// Simple in-memory rate limiter per IP
type bucket struct {
	tokens float64
	last   time.Time
}

var rl = make(map[string]*bucket)
var rlMu sync.Mutex

func allow(ip string) bool {
	rlMu.Lock()
	defer rlMu.Unlock()
	b, ok := rl[ip]
	if !ok {
		b = &bucket{tokens: 10, last: time.Now()}
		rl[ip] = b
	}
	elapsed := time.Since(b.last).Seconds()
	b.tokens += elapsed * 1.0 // 1 token/s refill
	if b.tokens > 10 {
		b.tokens = 10
	}
	b.last = time.Now()
	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

type WafProxy struct {
	target string
	proxy  *httputil.ReverseProxy
}

func newWafProxy(target string) *WafProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = target
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
		ModifyResponse: func(resp *http.Response) error {
			return nil
		},
	}

	return &WafProxy{
		target: target,
		proxy:  proxy,
	}
}

func (wp *WafProxy) checkSQLInjection(r *http.Request, ip string) error {
	// Check URL query parameters
	if err := checkMaliciousContent(r.URL.RawQuery); err != nil {
		log.Printf("Blocked malicious attempt from %s in query: %v", ip, err)
		return fmt.Errorf("malicious content detected in query: %v", err)
	}

	// Check POST body
	if r.Method == "POST" {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		// Check raw body for malicious content
		if err := checkMaliciousContent(string(body)); err != nil {
			log.Printf("Blocked malicious attempt from %s in body: %v", ip, err)
			return fmt.Errorf("malicious content detected in body: %v", err)
		}

		contentType := r.Header.Get("Content-Type")
		log.Printf("Content-Type %s", contentType)

		// Handle form-encoded data
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// Create a copy of the body for parsing
			bodyBytes := make([]byte, len(body))
			copy(bodyBytes, body)
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			// Parse the form data
			err := r.ParseForm()
			if err == nil {
				// Check form fields
				for field, values := range r.Form {
					for _, value := range values {
						log.Printf("Checking form field %s value: %s", field, value)
						if err := checkMaliciousContent(value); err != nil {
							log.Printf("Blocked malicious attempt from %s in form field %s: %v", ip, field, err)
							return fmt.Errorf("malicious content detected in form field %s: %v", field, err)
						}
					}
				}
			}
			// Reset the body to the original content
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		// Handle JSON data
		if strings.Contains(contentType, "application/json") {
			var jsonBody map[string]interface{}
			if json.Unmarshal(body, &jsonBody) == nil {
				// Recursively check all string values in JSON
				var checkJSON func(interface{}) error
				checkJSON = func(v interface{}) error {
					switch val := v.(type) {
					case string:
						return checkMaliciousContent(val)
					case map[string]interface{}:
						for _, value := range val {
							if err := checkJSON(value); err != nil {
								return err
							}
						}
					case []interface{}:
						for _, item := range val {
							if err := checkJSON(item); err != nil {
								return err
							}
						}
					}
					return nil
				}

				if err := checkJSON(jsonBody); err != nil {
					log.Printf("Blocked malicious attempt from %s in JSON: %v", ip, err)
					return fmt.Errorf("malicious content detected in JSON: %v", err)
				}
			}
		} // Recreate body for upstream
		r.Body = io.NopCloser(bytes.NewReader(body))
	}
	return nil
}

func (wp *WafProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	// Check rate limit
	if !allow(ip) {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	// Check for SQL injection
	if err := wp.checkSQLInjection(r, ip); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Forward the request
	wp.proxy.ServeHTTP(w, r)
}

func main() {
	target := "127.0.0.1:8080" // upstream app
	http.ListenAndServe(":8081", newWafProxy(target))
}
