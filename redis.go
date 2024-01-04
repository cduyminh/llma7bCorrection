package main

import (
	"context"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	once     sync.Once
	instance *redis.Client
)

// GetRedisClient returns a singleton instance of the Redis client.
func GetRedisClient() *redis.Client {
	withTLS := true

	if Cfg.RedisTLS == "false" {
		withTLS = false
	}

	// Use sync.Once to ensure the initialization code is executed only once.
	once.Do(func() {
		if withTLS {
			log.Println("Connecting to Redis server via TLS")
			instance = connectToRedisViaTLS()
		} else {
			log.Println("Connecting to Redis server via Localhost")
			instance = createRedisClient()
		}

		if err := instance.Ping(context.Background()).Err(); err != nil {
			log.Println("Failed to connect to Redis server")
		} else {
			log.Println("Connected to Redis server")
		}
	})

	return instance
}

// createRedisClient creates and configures a new Redis client.
func createRedisClient() *redis.Client {
	// Configure the Redis client with your Redis server information.
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6370", // Change this to your Redis server address.
		Password: "",               // No password by default.
		DB:       0,                // Default DB.
	})

	return client
}

func connectToRedisViaTLS() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     Cfg.RedisAdrr, // Change this to your Redis server address.
		Username: Cfg.RedisUsername,
		Password: Cfg.RedisPassword, // No password by default.
		DB:       0,                 // Default DB.

	})

	return client
}
