package main
import (
  "container/list"
  "encoding/json"
  "errors"
  "reflect"
  "fmt"
  "github.com/bradfitz/go-smtpd/smtpd"
  "log"
  "net/http"
  "strings"
  "sync"
  "flag"
  //"reflect"
)

type List struct {
  root Element // sentinel list element, only &root, root.prev, and root.next are used
  len  int     // current list length excluding (this) sentinel element
}

type Message struct {
  From string
  To []string
  Body string
}

var l = list.New()
var mutex = &sync.Mutex{}

var myMessage Message
var localcount int32 = 0

type env struct {
  *smtpd.BasicEnvelope
  msg Message
}


type Element struct {
  next, prev *Element
  // The list to which this element belongs.
  list *List
  // The value stored with this element.
  Value Message
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
  myMessage.From = from.Email()
  return &env{new(smtpd.BasicEnvelope), myMessage}, nil
}

func (e *env) AddRecipient(rcpt smtpd.MailAddress) error {
  if strings.HasPrefix(rcpt.Email(), "bad@") {
    return errors.New("we don't send email to bad@")
  }
  fmt.Println("Typeof to list", reflect.TypeOf(e.msg.To))
  e.msg.To = append(e.msg.To, rcpt.Email())
  fmt.Println(e.msg.To)
  // e.msg.To[localcount] = rcpt.Email()
  // localcount++
  return e.BasicEnvelope.AddRecipient(rcpt)
}

func (e *env) Write(line []byte) error {
  log.Printf("Line: %q", string(line))
  str := strings.Replace(string(line),"\n"," ",-1)
  str = strings.Replace(str,"\r"," ",-1)
  e.msg.Body += str
  return nil
}

func (e *env) Close() error  {
  log.Printf("The message: %q", e.msg)
  localcount = 0
  mutex.Lock()
  l.PushBack(e.msg)
  mutex.Unlock()
  return nil
}

func retrieve_mail(w http.ResponseWriter, req *http.Request) {
  fmt.Println("*******in handler get**********")

  w.Header().Set("Server", "A Go Web Server")
  w.WriteHeader(200)
  mutex.Lock()
  for e := l.Front(); e != nil; e = e.Next() {
    if e.Value.(Message).From == req.FormValue("From"){
      fmt.Println("Value removed from queue:", e.Value.(Message))
      out, err :=json.MarshalIndent(e.Value.(Message), "", "    ")
      if err != nil {
        fmt.Println("error:", err)
      }
      w.Write([]byte(out))
    }
  }
  mutex.Unlock()
}

func retrieve_all(w http.ResponseWriter, req *http.Request) {
  mutex.Lock()
  for e := l.Front(); e != nil; e = e.Next() {

    value := e.Value.(Message)
    out, err :=json.MarshalIndent(value, "", "    ")
    if err != nil {
      fmt.Println("error:", err)
    }
    w.Header().Set("Server", "A Go Web Server")
    w.WriteHeader(200)
    w.Write([]byte(out))
    w.Write([]byte("\n"))


  }
mutex.Unlock()
  }

func pop(w http.ResponseWriter, req *http.Request) {
  fmt.Println("******in handler**********")
  mutex.Lock()
  first_mail := l.Front()
  value := l.Remove(first_mail)
  fmt.Println("Value removed from queue:", value)
  mutex.Unlock()
  out, err :=json.MarshalIndent(value.(Message), "", "    ")
  if err != nil {
    fmt.Println("error:", err)
  }
  w.Header().Set("Server", "A Go Web Server")
  w.WriteHeader(200)
  w.Write([]byte(out))
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
  http.HandleFunc("/pop", pop)
  http.HandleFunc("/mails", retrieve_all)
  http.HandleFunc("/mails/get", retrieve_mail)

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
