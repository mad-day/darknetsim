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


package client

import (
	"github.com/xtaci/smux"
	
	//"io"
	"net"
	"errors"
)

var EBadArg = errors.New("Bad Args")
var EShort = errors.New("Short")

// The client is a listener that offers a hidden service (.onion) within the simulated darknet.
type Client struct {
	sess *smux.Session
}

func NewClient(c net.Conn, addr string) (cli *Client,err error) {
	if len(addr)>128 { err = EBadArg; return }
	buf := make([]byte,128)
	copy(buf,addr)
	n,_ := c.Write(buf)
	if n!=128 { err = EShort; return }
	cfg := smux.DefaultConfig()
	sms,e := smux.Client(c,cfg)
	if e!=nil { err = e; return }
	return &Client{sms}, nil
}

var _ net.Listener = (*Client)(nil)

func (c *Client) Accept() (nc net.Conn, err error) {
	nc,err = c.sess.AcceptStream()
	if err!=nil { nc = nil }
	return
}

func (c *Client) Close() error { return c.sess.Close() }
func (c *Client) Addr() net.Addr { return c.sess.LocalAddr() }

