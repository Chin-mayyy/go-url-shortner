package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type ShorternRequest struct {
	URL string `json:"url"`
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func main() {
	redisUrl := os.Getenv("REDIS_URL")
	opt, _ := redis.ParseURL(redisUrl)
	rdb := redis.NewClient(opt)

	router := http.NewServeMux()

	router.HandleFunc("GET /{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		url := rdb.Get(ctx, slug).Val()
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	})

	router.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		var body ShorternRequest
		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil {
			fmt.Println("Error Decoding: ", err.Error())
			return
		}

		shortKey := generateShortKey()
		rdb.Set(ctx, shortKey, body.URL, 0)

		w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", shortKey)))
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Server listening on port :8080")
	server.ListenAndServe()
}
