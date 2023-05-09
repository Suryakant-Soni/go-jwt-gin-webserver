package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getMongoClient() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading env file", err)
	}
	MongoDbUrl := os.Getenv("MONGODB_URL")
	// getting the mongo client in hand for that particular url
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDbUrl))
	if err != nil {
		log.Fatal("Error", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Error", err)
	}
	log.Println("connected to mongo db modified")
	return client
}

var Client *mongo.Client = getMongoClient()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("mymongo").Collection(collectionName)
	return collection
}
