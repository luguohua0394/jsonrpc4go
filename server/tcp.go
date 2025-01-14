package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sunquakes/jsonrpc4go/common"
	"github.com/sunquakes/jsonrpc4go/discovery"
	"golang.org/x/time/rate"
	"log"
	"net"
	"sync"
)

type Tcp struct {
	Hostname string
	Port     int
}

type TcpServer struct {
	Hostname  string
	Port      int
	Server    common.Server
	Options   TcpOptions
	Event     chan int
	Discovery discovery.Driver
}

type TcpOptions struct {
	PackageEof       string
	PackageMaxLength int64
}

func (p *Tcp) NewServer() Server {
	options := TcpOptions{
		"\r\n",
		1024 * 1024 * 2,
	}
	var err error
	if p.Hostname == "" {
		p.Hostname, err = GetHostname()
		if err != nil {
			log.Panic(err.Error())
		}
	}
	return &TcpServer{
		p.Hostname,
		p.Port,
		common.Server{
			sync.Map{},
			common.Hooks{},
			nil,
		},
		options,
		make(chan int, 1),
		nil,
	}
}

func (s *TcpServer) Start() {
	// Register services
	if s.Discovery != nil {
		register := func(key, value interface{}) bool {
			err := s.Discovery.Register(key.(string), "tcp", s.Hostname, s.Port)
			if err == nil {
				return true
			}
			return false
		}
		s.Server.Sm.Range(register)
	}
	// Start the server
	var addr = fmt.Sprintf("0.0.0.0:%d", s.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Panic(err.Error())
	}
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	log.Printf("Listening tcp://0.0.0.0:%d", s.Port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Event <- 0
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Panic(err.Error())
			continue
		}
		go s.handleFunc(ctx, conn)
	}
}

func (s *TcpServer) Register(m any) {
	s.Server.Register(m)
}

func (s *TcpServer) SetOptions(tcpOptions any) {
	s.Options = tcpOptions.(TcpOptions)
}

func (s *TcpServer) SetDiscovery(d discovery.Driver) {
	s.Discovery = d
}

func (s *TcpServer) SetRateLimit(r rate.Limit, b int) {
	s.Server.RateLimiter = rate.NewLimiter(r, b)
}

func (s *TcpServer) SetBeforeFunc(beforeFunc func(id any, method string, params any) error) {
	s.Server.Hooks.BeforeFunc = beforeFunc
}

func (s *TcpServer) SetAfterFunc(afterFunc func(id any, method string, result any) error) {
	s.Server.Hooks.AfterFunc = afterFunc
}

func (s *TcpServer) GetEvent() <-chan int {
	return s.Event
}

func (s *TcpServer) handleFunc(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	select {
	case <-ctx.Done():
		return
	default:
		//	do nothing
	}
	eofb := []byte(s.Options.PackageEof)
	eofl := len(eofb)
	for {
		var (
			data []byte
		)
		l := 0
		for {
			var buf = make([]byte, s.Options.PackageMaxLength)
			n, err := conn.Read(buf)
			if err != nil {
				if n == 0 {
					return
				}
				common.Debug(err.Error())
			}
			l += n
			data = append(data, buf[:n]...)
			if bytes.Equal(data[l-eofl:], eofb) {
				break
			}
		}
		res := s.Server.Handler(data[:l-eofl])
		res = append(res, eofb...)
		conn.Write(res)
	}
}
