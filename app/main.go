package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq" // Driver สำหรับ Postgres
)

var db *sql.DB

func main() {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		log.Printf("⏳ Waiting for database... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("❌ Could not connect to database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
		short_key TEXT PRIMARY KEY,
		original_url TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// ใน main.go แก้ไขส่วนนี้
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// ถ้าขอ /r/ หรือ /shorten จะไปเข้า Handler อื่นเอง
			// แต่ถ้าขอมั่วๆ เช่น /abc ให้ส่ง 404 ของ Go ปกติ
			http.NotFound(w, r)
			return
		}
		// ใช้เครื่องหมาย ./ เพื่อบอกว่าเอาไฟล์ที่อยู่ข้างๆ ตัว binary เลย
		http.ServeFile(w, r, "./index.html")
	})

	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/r/", redirectHandler)

	log.Println("URL shortener with DB started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {

	mySecretKey := os.Getenv("API_KEY")
	if mySecretKey == "" {
		http.Error(w, "Server misconfigured: API_KEY is not set", http.StatusInternalServerError)
		return
	}
	clientKey := r.Header.Get("X-API-Key")

	if clientKey != mySecretKey {
		http.Error(w, "🚫 Unauthorized: Wrong API Key", http.StatusUnauthorized)
		return
	}

	key := r.URL.Query().Get("key")
	url := r.URL.Query().Get("url")

	if key == "" || url == "" {
		http.Error(w, "Missing key or url", 400)
		return
	}

	_, err := db.Exec("INSERT INTO urls (short_key, original_url) VALUES ($1, $2) ON CONFLICT (short_key) DO UPDATE SET original_url = $2", key, url)
	if err != nil {
		http.Error(w, "DB Error", 500)
		return
	}

	fmt.Fprintf(w, "✅ Shortened! Your link is: /r/%s", key)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/r/"):]
	var originalURL string

	err := db.QueryRow("SELECT original_url FROM urls WHERE short_key = $1", key).Scan(&originalURL)
	if err != nil {
		http.Error(w, "Link not found", 404)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
