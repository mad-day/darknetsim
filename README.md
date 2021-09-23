# darknetsim
Darknet Simulator to aid the developement of P2P applications in the darknet.

The Idea is to emulate the approximate behavoir of hidden services without really polluting TOR with test services.

## Getting started.

Get/compile the server with `go get github.com/mad-day/darknetsim/dnsim`.

Get the client library with `go get -u github.com/mad-day/darknetsim/client`.

This library is used to create "hidden services" in this simulation/fake darknet.


## Hidden service example:

```go
package main

import "github.com/mad-day/darknetsim/client"

import "log"
import "net"
import "net/http"
import "fmt"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	})
	c,e := net.Dial("tcp","127.0.0.1:9996")
	if e!=nil { log.Fatal(e) }
	
	li,e := client.NewClient(c,"6sxoyfb3h2nvok2d.onion:80")
	if e!=nil { log.Fatal(e) }
	e = http.Serve(li,http.DefaultServeMux)
	if e!=nil { log.Fatal(e) }
}
```

## Client example:

```go
package main

import (
	"golang.org/x/net/proxy"
	"log"
	"net/http"
	"os"
	"io"
)

var dia proxy.Dialer

func main() {
	var err error
	dia,err = proxy.SOCKS5("tcp","127.0.0.1:9991",nil,nil)
	if err!=nil { log.Fatal(err) }
	
	tp := new(http.Transport)
	
	if cdia,ok := dia.(proxy.ContextDialer); ok {
		tp.DialContext = cdia.DialContext
	} else {
		tp.Dial = dia.Dial
	}
	cl := &http.Client{ Transport: tp }
	
	resp,err := cl.Get("http://6sxoyfb3h2nvok2d.onion/Hello_World")
	if err!=nil { log.Fatal(err) }
	
	io.Copy(os.Stdout,resp.Body)
	resp.Body.Close()
	os.Stdout.Write([]byte("\r\n"))
	//req := http.NewRequest("GET", "http://6sxoyfb3h2nvok2d.onion/Hello_World", nil)
}
```
