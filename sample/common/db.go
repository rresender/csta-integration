package db

import (
	"log"
	"time"

	"github.com/go-redis/redis"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//Conenct to Redis
func Connect(redisHost string) *redis.Client {

	db := redis.NewClient(&redis.Options{
		Addr:         redisHost,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	if err := db.Ping().Err(); err != nil {
		FailOnError(err, "Failed to connect to redis")
	}

	log.Println("Redis connected...")

	return db
}
