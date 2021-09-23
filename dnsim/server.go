/*
Copyright (c) 2021 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package main


import "github.com/mad-day/darknetsim/serverlib"
import flags "flag"
import "net"
import "log"

var cli_prt = flags.String("c", "127.0.0.1:9991", "socks5 proxy port")
var srv_prt = flags.String("s", "127.0.0.1:9996", "smux service port")
var help = flags.Bool("h",false,"Help!")

var finished = make(chan int,16)
func signl() { finished <- 1 }

func main(){
	flags.Parse()
	if *help {
		flags.PrintDefaults()
		return
	}
	
	cll,e := net.Listen("tcp4",*cli_prt)
	if e!=nil { log.Fatal(e) }
	defer cll.Close()
	sll,e := net.Listen("tcp4",*srv_prt)
	if e!=nil { return }
	defer sll.Close()
	
	var srv = new(serverlib.Server)
	
	go func(){
		defer signl()
		for {
			c,e := cll.Accept()
			if e!=nil { continue }
			srv.ServeSocks5(c)
		}
	}()
	go func(){
		defer signl()
		for {
			c,e := sll.Accept()
			if e!=nil { continue }
			srv.ServeService(c)
		}
	}()
	<- finished
	<- finished
}

