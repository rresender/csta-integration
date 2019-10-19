package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	db "github.com/rresender/csta-integration/sample/common"
)

var conn *redis.Client
var port = ":7070"

func init() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis:6379"
	}
	conn = db.Connect(redisHost)
}

func main() {

	defer conn.Close()

	m := mux.NewRouter()

	m.HandleFunc("/callinfo/{ucid}", func(w http.ResponseWriter, r *http.Request) {

		host, _ := os.Hostname()
		log.Printf("request processed by host: %s\n", host)

		vars := mux.Vars(r)
		UCID := vars["ucid"]

		js := conn.Get(UCID)

		if js.Err() != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, fmt.Sprintf("No Call found for UCID: %s", UCID), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(js.Val()))
	})

	log.Printf("HTTP Server Listening at %s\n", port)
	log.Fatal(http.ListenAndServe(port, m))
}
