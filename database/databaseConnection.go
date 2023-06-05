package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client { //.env file -> wahi se db connection hoga
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	MongoDb := os.Getenv("MONGODB_URL") //.env file se mongodb ki url leke aao

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client

}

var Client *mongo.Client = DBinstance() //we call this fn -> it'll return us the mongo client-> and we have captured it in the var Client

//now we're going to use a fn to access a particular collection in the db
func OpenCollection(client *mongo.Client, collectionNmae string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("cluster0").Collection(collectionNmae)
	return collection
}
