package tcp

import (
	"fmt"
	"net"
)

type Server struct {
	listenAddress string
	ln            net.Listener
	quitCH        chan struct{}
}

func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		quitCH:        make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddress)

	if err != nil {
		panic(err)
	}

	defer ln.Close()
	s.ln = ln

	<-s.quitCH
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()

		if err != nil {
			fmt.Println("error accepting connection", err)
		}

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)

		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("client disconnected")
				break
			} else {
				fmt.Println("error reading from connection", err)
				continue
			}
		}

		msg := (buf[:n])
		fmt.Println("message received", string(msg))

	}
}
