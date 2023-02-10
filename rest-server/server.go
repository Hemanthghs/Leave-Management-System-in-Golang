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

type leave_item struct {
	Name   string `json:"name"`
	Id     int32  `json:"id"`
	Reason string `json:"reason"`
	Date   string `json:"date"`
	Status string `json:"status"`
}

type leave_approve struct {
	LeaveId int32  `json:"leaveId"`
	Status  string `json:"status"`
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
var leave_collection *mongo.Collection

func connectToDB() {
	mongo_uri := goDotEnvVariable("MONGODB_URI")
	client, err := mongo.NewClient(options.Client().ApplyURI(mongo_uri))
	handleError(err)
	fmt.Println("DB Connected...")
	err = client.Connect(context.TODO())
	handleError(err)
	students_collection = client.Database("lms").Collection("students")
	leave_collection = client.Database("lms").Collection("leaves")
}

// this function is used to add new student, admin can add the students
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

// this function is request the leave, student can use this function
func leaveRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var leave leave_item
	json.NewDecoder(r.Body).Decode(&leave)
	filter := bson.M{
		"id": leave.Id,
	}
	var result_data []student_item
	cursor, err := students_collection.Find(context.TODO(), filter)
	cursor.All(context.Background(), &result_data)
	handleError(err)
	if len(result_data) == 0 {
		w.Write([]byte("Student does not exist"))
		return
	}
	leave.Status = "Pending"
	leave_collection.InsertOne(context.TODO(), leave)
	w.Write([]byte("Leave request sent successfully"))
}

// this function is used to list all the leaves, admin can use this
func viewLeaves(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var leaves []leave_item
	cursor, err := leave_collection.Find(context.TODO(), bson.D{{}})
	handleError(err)
	cursor.All(context.Background(), &leaves)
	fmt.Println(leaves)
	for _, val := range leaves {
		s := fmt.Sprintf("%s %d %s %s %s\n", val.Name, val.Id, val.Reason, val.Date, val.Status)
		w.Write([]byte(s))
	}
}

// this function is used to accept or reject a leave, the admin can provide the leaveId and the status
func approveLeave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var leaveApprove leave_approve
	json.NewDecoder(r.Body).Decode(&leaveApprove)
	filter := bson.M{
		"leaveId": leaveApprove.LeaveId,
	}

	update := bson.D{{"$set", bson.D{{"status", leaveApprove.Status}}}}
	_, err := leave_collection.UpdateOne(context.TODO(), filter, update)
	handleError(err)
	w.Write([]byte("Status updated"))

}

// this function is used to check the status of the leave, student can use this
func checkStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var student student_item
	json.NewDecoder(r.Body).Decode(&student)
	filter := bson.M{
		"id": student.Id,
	}
	var result_data []leave_item
	cursor, err := leave_collection.Find(context.TODO(), filter)
	cursor.All(context.Background(), &result_data)
	handleError(err)
	if len(result_data) == 0 {
		w.Write([]byte("No leave request found"))
		return
	}
	result := fmt.Sprint("Status ", result_data[0].Status)
	w.Write([]byte(result))
}

// this function is used to initialize the routes
func initializeRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/addstudent", addStudent).Methods("POST")
	r.HandleFunc("/leaverequest", leaveRequest).Methods("POST")
	r.HandleFunc("/checkstatus", checkStatus).Methods("POST")
	r.HandleFunc("/viewleaves", viewLeaves).Methods("GET")
	r.HandleFunc("/approveleave", approveLeave).Methods("POST")
	log.Fatal(http.ListenAndServe("0.0.0.0:10000", r))
}

func main() {
	connectToDB()
	initializeRouter()
}
