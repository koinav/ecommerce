package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var Client = DBSetup()

func DBSetup() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://development:testpassword@localhost:27017"))
	if err != nil {
		log.Println("failed to connect to mongodb")
		panic(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("failed to connect to mongodb")
		panic(err)
	}

	fmt.Println("Successfully connected to MongoDB")
	return client
}

func UserData(client *mongo.Client, collectionName string) *mongo.Collection {
	var userCollection = client.Database("Ecommerce").Collection(collectionName)

	return userCollection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection {
	var productCollection = client.Database("Ecommerce").Collection(collectionName)

	return productCollection
}
