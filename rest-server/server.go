package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type student_item struct {
	Name string `json:"name"`
	Id   int32  `json:"id"`
}

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
}

func addStudent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var student student_item
	json.NewDecoder(r.Body).Decode(&student)
	id := student.Id
	filter := bson.M{
		"id": id,
	}
	var result_data []student_item
	cursor, err := students_collection.Find(context.TODO(), filter)
	cursor.All(context.Background(), &result_data)
	handleError(err)
	if len(result_data) != 0 {
		w.Write([]byte("Student Already Exist"))
		return
	}
	students_collection.InsertOne(context.TODO(), student)
	w.Write([]byte("Student details added"))
}

func initializeRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/addstudent", addStudent).Methods("POST")
	log.Fatal(http.ListenAndServe("0.0.0.0:10000", r))
}

func main() {
	connectToDB()
	initializeRouter()
}
