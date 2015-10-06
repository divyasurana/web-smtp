package main

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bradfitz/go-smtpd/smtpd"
	"log"
	"net/http"
	"strings"
	"sync"
	"flag"
)

type Message struct {
	From string     `json:"from"`
	To   [20]string `json:"to"`
	Body string     `json:"body"`
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
	log.Printf("The message: %q", myMessage)
	localcount = 0
	mutex.Lock()
	l.PushBack(e.msg)
	mutex.Unlock()
	return nil
}

func retrieve_mail(w http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	first_mail := l.Front()
	value := l.Remove(first_mail)
	fmt.Println("Value removed from queue:", value)
	mutex.Unlock()

	res1B, _ := json.Marshal(value)
	w.Header().Set("Server", "A Go Web Server")
	w.WriteHeader(200)
	w.Write([]byte(res1B))
}

func SMTPListener(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)

	s := &smtpd.Server{
		Hostname:  host,
		Addr:     addr,
		OnNewMail: onNewMail,
	}

	log.Println("[SMTP] Starting SMTP interface on", addr)

	err := s.ListenAndServe()

	if err != nil {
		log.Fatal("[SMTP] ERROR:", err)
	}
}

func HTTPListener(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	http.HandleFunc("/mails", retrieve_mail)

	log.Println("[HTTP] Starting HTTP interface on", addr)

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("[HTTP] ERROR:", err)
	}
}

var (
	host = flag.String("b", "0.0.0.0", "listen on HOST")
	httpPort = flag.Int("p", 8080, "use PORT for HTTP")
	smtpPort = flag.Int("s", 587, "use PORT for SMTP")
)

func main() {
	flag.Parse()

	go SMTPListener(*host, *smtpPort)
	HTTPListener(*host, *httpPort)
}
