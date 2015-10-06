package main

import (
	"log"
	"fmt"
	"net/http"
	"errors"
	"sync"
	"encoding/json"
	"container/list"
	"strings"
	"github.com/bradfitz/go-smtpd/smtpd"
)

type Message struct {
	From  string	`json:"from"`
	To [20]string `json:"to"`
	Body     string `json:"body"`
}

var l = list.New()
var mutex = &sync.Mutex{}

var myMessage Message
var localcount int32 = 0

type env struct {
	*smtpd.BasicEnvelope
	msg Message
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
  myMessage.From = from.Email()
	return &env{new(smtpd.BasicEnvelope), myMessage}, nil
}

func (e *env) AddRecipient(rcpt smtpd.MailAddress) error {
	if strings.HasPrefix(rcpt.Email(), "bad@") {
		return errors.New("we don't send email to bad@")
	}
	e.msg.To[localcount] = rcpt.Email()
	localcount++
	return e.BasicEnvelope.AddRecipient(rcpt)
}

func (e *env) Write(line []byte) error {
	log.Printf("Line: %q", string(line))
	e.msg.Body += string(line)
	return nil
}

func (e *env) Close() error {
	log.Printf("The message: %q",myMessage)
	localcount = 0
	mutex.Lock()
	l.PushBack(e.msg)
	mutex.Unlock()
	return nil
}

func retrieve_mail(w http.ResponseWriter, req *http.Request) {
  fmt.Println("******in handler**********")
	mutex.Lock()
	first_mail := l.Front()
	value := l.Remove(first_mail)
	fmt.Println("Value removed from queue:", value)
	mutex.Unlock()

	// res1D := &Message{
  //       From:   value.From,
  //       To: []string{value.To[0]},
	// 			Body: value.Body,
	// 			}
	res1B, _ := json.Marshal(value)
	w.Header().Set("Server", "A Go Web Server")
  w.WriteHeader(200)
	w.Write([]byte(res1B))
}

func smtp_listen() {
	s := &smtpd.Server{
	      Hostname: "172.18.17.20",
		  	Addr:":2500",
		 		OnNewMail: onNewMail,
	}
	err1 := s.ListenAndServe()
	if err1 != nil {
		log.Fatalf("ListenAndServe: %v", err1)
	}
}

func http_listen() {
	http.HandleFunc("/mails", retrieve_mail)
	err := http.ListenAndServe(":8080",nil)
	if err != nil {
			log.Fatal("ListenAndServe:", err)
	}
}

func main() {

	fmt.Println("******in main**********")
	go smtp_listen()
	go http_listen()

	var input string
    fmt.Scanln(&input)
    fmt.Println("done")
}

