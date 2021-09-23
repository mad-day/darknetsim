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


package serverlib

import (
	"github.com/xtaci/smux"
	socks5 "github.com/armon/go-socks5"
	"io"
	"net"
	"sync"
	
	// yes... legacy
	"golang.org/x/net/context"
	"errors"
)

var mutex sync.Mutex

func dolock() func() {
	mutex.Lock()
	return mutex.Unlock
}

type Server struct{
	listens sync.Map
	cfg *smux.Config
	srv *socks5.Server
}
var notFound = errors.New("notfound")

func kick(s *Server,str string) bool {
	defer dolock()()
	val,ok := s.listens.Load(str)
	if !ok { return true }
	if !val.(*service).sess.IsClosed() { return false }
	s.listens.Delete(str)
	return true
}

func handshake(s *Server,c net.Conn) {
	buf := make([]byte,128)
	_,err := io.ReadFull(c,buf)
	if err!=nil { c.Close(); return }
	l := len(buf)
	for i,b := range buf { if b==0 { l = i; break } }
	str := string(buf[:l])
	if s.cfg==nil { s.cfg = smux.DefaultConfig() }
	fc := &filterSocket{c,nil}
	sms,err := smux.Server(fc,s.cfg)
	fc.sess = sms
	if err!=nil { c.Close(); return }
	svc := &service{ sms, c }
	for {
		_,cnf := s.listens.LoadOrStore(str,svc)
		if !cnf { break }
		if kick(s,str) { continue }
		sms.Close()
		break
	}
}

type filterSocket struct{
	net.Conn
	sess *smux.Session
}
func (fs *filterSocket) Read(buf []byte) (int,error) {
	i,err := fs.Conn.Read(buf)
	if err!=nil {
		s := fs.sess
		fs.sess = nil
		if s!=nil { s.Close() }
	}
	return i,err
}

type service struct{
	sess *smux.Session
	sock net.Conn
}


type fakeSock struct{
	*smux.Stream
}
func (fs *fakeSock) CloseWrite() error { return fs.Close() }


func (s *Server) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	raw,ok := s.listens.Load(addr)
	if !ok { return nil,notFound }
	c,err := raw.(*service).sess.OpenStream()
	if err!=nil { return nil,err }
	return &fakeSock{c},nil
}

func (s *Server) ServeService(c net.Conn) { handshake(s,c) }

type impler int
func (impler) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) { return ctx,nil,nil }


var s5cfg = &socks5.Config{
	Resolver: impler(0),
}

func (s *Server) makesvc() (err error) {
	if s.srv!=nil { return }
	cfg := new(socks5.Config)
	*cfg = *s5cfg
	cfg.Dial = s.dial
	s.srv,err = socks5.New(cfg)
	return
}

func (s *Server) ServeSocks5(c net.Conn) {
	err := s.makesvc()
	if err!=nil { c.Close() }
	s.srv.ServeConn(c)
}

