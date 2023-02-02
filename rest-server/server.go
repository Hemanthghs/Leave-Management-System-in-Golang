package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	handleError(err)
	return os.Getenv(key)
}

var students_collection *mongo.Collection

func connectToDB() {
	mongo_uri := goDotEnvVariable("MONGODB_URI")
	client, err := mongo.NewClient(options.Client().ApplyURI(mongo_uri))
	handleError(err)
	fmt.Println("DB Connected...")
	err = client.Connect(context.TODO())
	handleError(err)
	students_collection = client.Database("lms").Collection("students")
	fmt.Println(students_collection)
}

func main() {
	connectToDB()
}
