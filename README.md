# web-smtp
A hybrid Go SMTP-HTTP server to support the storing of incoming mails and retrieval through a REST API.

# Installation

```
$ go get https://github.com/divyasurana/web-smtp.git
```

# Usage
Starting the server
```
Usage of ./web-smtp:
  -b string
    	listen on HOST (default "0.0.0.0")
  -p int
    	use PORT for HTTP (default 8080)
  -s int
    	use PORT for SMTP (default 587)
```

Fetching mails using curl

```
$ curl http://localhost:8080/mails

{
  "from": "Alice",
  "to": ["Bob"],
  "body": "mail body ...\r\n"
}
{
  "from": "Alice",
  "to": ["John"],
  "body": "mail body ...\r\n"
}

{
  "from": "Bob",
  "to": ["John"],
  "body": "mail body ...\r\n"
}
```
```
$ curl http://localhost:8080/pop
{
  "from": "Alice",
  "to": ["Bob"],
  "body": "mail body ...\r\n"
}

```

```
$ curl http://localhost:8080/mails/get?From=Alice
{
  "from": "Alice",
  "to": ["Bob"],
  "body": "mail body ...\r\n"
}
{
  "from": "Alice",
  "to": ["John"],
  "body": "mail body ...\r\n"
}
```
