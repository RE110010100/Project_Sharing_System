package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	UserID        string         `bson:"userid"`
	Name          string         `bson:"name"`
	Notifications []Notification `bson:"notifications"`
}

type Notification struct {
	Message string `bson:"message"`
	Time    string `bson:"time"`
}

func connectToMongoDB() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

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

func AddUserNotificationToDB(client *mongo.Client) (*mongo.UpdateResult, error) {

	// Specify the database and collection
	db := client.Database("NotificationsDB")
	collection := db.Collection("UserNotifications")

	// Specify filter and update
	filter := bson.M{"userid": "email1"} // Filter to identify the document to update
	update := bson.M{
		"$push": bson.M{
			"notifications": bson.M{
				"message": "Your new message",
				"time":    time.Now().Format(time.RFC3339),
			},
		},
	}

	// Perform update operation
	updateResult, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}

	fmt.Println("Matched:", updateResult.MatchedCount, "Modified:", updateResult.ModifiedCount)

	return updateResult, err

}

func AddUserDocumentToDB(client *mongo.Client, newUser *User) (*mongo.InsertOneResult, error) {

	// Specify the database and collection
	db := client.Database("NotificationsDB")
	collection := db.Collection("UserNotifications")

	// Insert the document into the collection
	insertResult, err := collection.InsertOne(context.Background(), newUser)
	if err != nil {
		return nil, err
	}

	fmt.Println("Inserted document ID:", insertResult.InsertedID)
	return insertResult, err
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

func NewUserWithEmptyNotifications(name string, userID string) *User {
	return &User{
		UserID:        userID,
		Name:          name,
		Notifications: make([]Notification, 0),
	}
}

func main() {

	//connect to MongoDB
	client := connectToMongoDB()

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

	/*_, err = AddUserNotificationToDB(client)
	if err != nil {
		fmt.Print(err)
	}*/

	user := NewUserWithEmptyNotifications("Rohan", "email1")

	_, err = AddUserDocumentToDB(client, user)
	if err != nil {
		fmt.Println(err)
	}

}
