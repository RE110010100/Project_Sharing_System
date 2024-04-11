package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service represents the notification service.
type Service struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

type Message struct {
	Text   string `json:"text"`
	UserID string `json:"userid"`
}

// NewService creates a new instance of the notification service.
func NewService(redisClient *redis.Client, mongoClient *mongo.Client) *Service {
	return &Service{redisClient: redisClient, mongoClient: mongoClient}
}

type User struct {
	UserID        string         `bson:"userid"`
	Name          string         `bson:"name"`
	Notifications []Notification `bson:"notifications"`
}

type Notification struct {
	Message string `bson:"message"`
	Time    string `bson:"time"`
}

// SubscribeToRedisChannel subscribes to a Redis channel and handles incoming messages.
func (s *Service) SubscribeToRedisChannel(channel string) {
	// Subscribe to the Redis channel
	pubsub := s.redisClient.PSubscribe(channel)
	defer pubsub.Close()

	// Channel to receive subscription messages
	ch := pubsub.Channel()

	// Handle messages asynchronously
	go func() {
		for msg := range ch {

			var message Message

			// Process incoming message
			fmt.Println(msg.Payload)
			err := json.Unmarshal([]byte(msg.Payload), &message)
			if err != nil {
				log.Printf("Failed to decode message: %v", err)
				continue
			}
			fmt.Printf("Received message: %s\n", message.Text)
			// Here, you can implement your notification logic

			_, err = AddUserNotificationToDB(s.mongoClient, message.Text, message.UserID)
			if err != nil {
				fmt.Print(err)
			}
		}
	}()

	// Block indefinitely (or until program termination)
	select {}
}

func AddUserNotificationToDB(client *mongo.Client, message, userID string) (*mongo.UpdateResult, error) {

	// Specify the database and collection
	db := client.Database("NotificationsDB")
	collection := db.Collection("UserNotifications")

	// Specify filter and update
	filter := bson.M{"userid": userID} // Filter to identify the document to update
	update := bson.M{
		"$push": bson.M{
			"notifications": bson.M{
				"message": message,
				"time":    time.Now().Format(time.RFC3339),
			},
		},
	}

	// Perform update operation
	updateResult, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}

	fmt.Println("Matched:", updateResult.MatchedCount, "Modified:", updateResult.ModifiedCount)

	return updateResult, err

}

func connectToMongoDB() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://admin:password123@mongodb:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return client
}

func CheckIndexExists(collection *mongo.Collection, keys bson.D) (bool, error) {
	ctx := context.Background()
	indexName := ""
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var index bson.M
		if err := cursor.Decode(&index); err != nil {
			return false, err
		}
		if key, ok := index["key"].(bson.M); ok {
			// Marshal keys to BSON bytes
			bsonKeys, err := bson.Marshal(keys)
			if err != nil {
				return false, err
			}
			// Unmarshal BSON bytes to bson.M
			var bsonM bson.M
			err = bson.Unmarshal(bsonKeys, &bsonM)
			if err != nil {
				return false, err
			}
			// Compare bson.M using reflect.DeepEqual
			if reflect.DeepEqual(key, bsonM) {
				indexName = index["name"].(string)
				break
			}
		}
	}
	if indexName != "" {
		return true, nil
	}
	return false, nil
}

// Create index function
func CreateIndexes(collection *mongo.Collection) error {
	indexOptions := options.Index().SetUnique(true)
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"userid": 1},
		Options: indexOptions,
	}
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	return err
}

func CreateNecessaryIndex(client *mongo.Client) {

	db := client.Database("NotificationsDB")
	collection := db.Collection("UserNotifications")

	exists, err := CheckIndexExists(collection, bson.D{{Key: "userid", Value: 1}})
	if err != nil {
		fmt.Println("error while checking")
		fmt.Println(err)
	}

	fmt.Println(exists)

	if !exists {
		// Create indexes
		err = CreateIndexes(collection)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Service) getUserNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse userID from query parameters
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	// Get a handle for the "Notifications" database and the "usernotifications" collection
	collection := s.mongoClient.Database("NotificationsDB").Collection("UserNotifications")

	// Define the filter to find documents with the specified userID
	filter := bson.M{"userid": userID}

	// Define a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the document that matches the filter
	var userNotification User
	err := collection.FindOne(ctx, filter).Decode(&userNotification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "No document found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to query MongoDB: %v", err), http.StatusInternalServerError)
		}
		return
	}

	fmt.Print(len(userNotification.Notifications))

	numNotifications := len(userNotification.Notifications)
	var notifications_tmp []Notification

	// Iterate over the last 10 elements of the notifications slice
	for i := numNotifications - 1; i >= numNotifications-10 && i >= 0; i-- {
		notifications_tmp = append(notifications_tmp, userNotification.Notifications[i])
	}

	// Marshal the notifications to JSON
	jsonResponse, err := json.Marshal(notifications_tmp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal notifications to JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// Set Content-Type header and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func main() {

	// Connect to Redis server.
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Redis server address.
		Password: "",           // No password set.
		DB:       0,            // Use default DB.
	})

	// Ping the Redis server to check if it's reachable.
	pong, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}
	fmt.Println("Connected to Redis:", pong)

	// Close the connection when done.
	defer client.Close()

	//connect to MongoDB
	mongoclient := connectToMongoDB()

	//create the necessary indicies
	CreateNecessaryIndex(mongoclient)

	// Create a new instance of the notification service
	notificationService := NewService(client, mongoclient)

	http.HandleFunc("/fetch_notifications", notificationService.getUserNotificationsHandler)

	// Start HTTP server
	go func() {
		if err := http.ListenAndServe(":8085", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Subscribe to the Redis channel
	notificationService.SubscribeToRedisChannel("channel.*")

	// Block indefinitely (or until program termination)
	select {}

}
