package main

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"fmt"
	"time"
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"io/ioutil"
)

type MessageStatus struct {
	MessagesSeen 	int					`json:"messagesSeen"`
	CurrentMessages int					`json:"currentMessages"`
	DeletedMessages int					`json:"deletedMessages"`
	Messages 		[]*MessageData		`json:"messages"`
}

type MessageData struct {
	ID 		*string
	Body 	*string
}

type ResponseMessage struct {
	Message string
}

var messageStatus *MessageStatus
var sqsClient *sqs.SQS
var qUrl = "https://sqs.eu-west-1.amazonaws.com/INSERT YOUR QUEUE URL HERE"


func messageScan() {
	fmt.Println("Scanning for messages....")
	for {
		result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{QueueUrl: &qUrl})
		if err == nil && len(result.Messages) > 0 {
			fmt.Println("Received message")
			messageStatus.Messages = append(messageStatus.Messages, &MessageData{result.Messages[0].MessageId, result.Messages[0].Body})
			messageStatus.CurrentMessages ++
			messageStatus.MessagesSeen ++
			fmt.Printf("There are now %v messages\n", messageStatus.CurrentMessages)
			sqsClient.DeleteMessage(&sqs.DeleteMessageInput{QueueUrl: &qUrl, ReceiptHandle: result.Messages[0].ReceiptHandle})
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/JSON")
	json.NewEncoder(w).Encode(messageStatus)
}

func getMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/JSON")
	vars := mux.Vars(r)
	messageId := vars["id"]
	found := false
	for _,v := range messageStatus.Messages {
		if *v.ID == messageId {
			json.NewEncoder(w).Encode(v)
			found = true
			break
		}
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
	}
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/JSON")
	vars := mux.Vars(r)
	messageId := vars["id"]
	for i,v := range messageStatus.Messages {
		if *v.ID == messageId {
			messageStatus.Messages = append(messageStatus.Messages[0:i], messageStatus.Messages[i+1:]...)
			messageStatus.CurrentMessages --
			messageStatus.DeletedMessages ++
			fmt.Println("Message deleted succesfully")
			break
		}
	}
	json.NewEncoder(w).Encode(messageStatus)
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/JSON")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
	} else {
		result, err := sqsClient.SendMessage(&sqs.SendMessageInput{
			MessageBody: aws.String(string(body)),
			QueueUrl:    &qUrl,
		})
		if err != nil {
			http.Error(w, "Problem sending message to SQS", http.StatusInternalServerError)
		} else {
			response := ResponseMessage{fmt.Sprint("Message successfully sent, ID:", *result.MessageId)}
			json.NewEncoder(w).Encode(response)
		}
	}
}

func status(w http.ResponseWriter, r *http.Request) {
	response := ResponseMessage{"Status OK"}
	json.NewEncoder(w).Encode(response)
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/status", status).Methods("GET")
	router.HandleFunc("/messages", getMessages).Methods("GET")
	router.HandleFunc("/messages/{id}", deleteMessage).Methods("DELETE")
	router.HandleFunc("/messages/{id}", getMessage).Methods("GET")
	router.HandleFunc("/messages", sendMessage).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func main() {
	sess := GetSession()
	sqsClient = sqs.New(sess)

	messageStatus = &MessageStatus{0,0,0, make([]*MessageData, 0)}

	go messageScan()

	startServer()
}
