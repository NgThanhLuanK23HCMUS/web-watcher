package main

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var templates = template.Must(template.ParseFiles(
	filepath.Join("templates", "home.html"),
	filepath.Join("templates", "login.html"),
	filepath.Join("templates", "captcha.html"),
))

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func captchaHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "captcha.html", nil)
}

func captchaVerifiedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie := http.Cookie{
		Name:     "captcha_verified",
		Value:    "true",
		MaxAge:   3000,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func requireCaptcha(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("captcha_verified")
		if err != nil || cookie.Value != "true" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Show login form for GET requests
		homeHandler(w, r)
	case http.MethodPost:
		// Handle login attempt for POST requests
		user := User{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}
		log.Printf("[LOGIN] %s attempted to log in with pw %s", user.Username, user.Password)

		fmt.Fprintf(w, "Hello, %s! You entered password: %s\n", html.EscapeString(user.Username), html.EscapeString(user.Password))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	log.Printf("[SEARCH] Query = %s", query)
	fmt.Fprintf(w, "<h2>Search results for: %s</h2>", html.EscapeString(query))
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339),
		"msg":  "Sample data from upstream server",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/", captchaHandler)
	http.HandleFunc("/verify-captcha", captchaVerifiedHandler)
	http.HandleFunc("/login", requireCaptcha(loginHandler))
	http.HandleFunc("/search", requireCaptcha(searchHandler))
	http.HandleFunc("/home", requireCaptcha(homeHandler))
	http.HandleFunc("/api/data", dataHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":8080"
	log.Printf("Upstream app listening on %s ...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
