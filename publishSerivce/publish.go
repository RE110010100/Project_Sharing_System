package main

import (
	"fmt"

	"github.com/go-redis/redis"
)

func publishMessageToRedis(redisClient *redis.Client, message, channel string) error {
	// Publish message to Redis channel
	err := redisClient.Publish(channel, message).Err()
	if err != nil {
		return err
	}
	return nil
}

func initializeRedisClient() *redis.Client {

	//var ctx = context.Background()

	// Initialize the Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Replace with your Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	// Ping the Redis server to check if it's reachable
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to Redis server")

	return client
}

func main() {

	// Initialize the Redis client
	redisClient := initializeRedisClient()
	defer redisClient.Close()

	// Publish message to Redis after successful upload
	err := publishMessageToRedis(redisClient, "Directory uploaded successfully!", "channel.upload")
	if err != nil {
		fmt.Print(err)
		return
	}

	// Publish message to Redis after successful upload
	err = publishMessageToRedis(redisClient, "Directory downloaded successfully!", "channel.download")
	if err != nil {
		fmt.Print(err)
		return
	}

}
