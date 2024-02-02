package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type Server struct {
	ListenAddress string
	Ln            net.Listener
	lockConn      sync.Mutex
	connections   map[string]*net.Conn
}

func (s *Server) startServer() {
	ln, err := net.Listen("tcp", s.ListenAddress)

	if err != nil {
		panic(err)
	}

	s.Ln = ln

	s.acceptLoop()
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.Ln.Accept()

		if err != nil {
			panic(err)
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
				// we also need to free up the hostname that was assigned to the client

				s.removeTheConnection(&conn)
				break
			} else {
				panic(err)
			}
		}

		reader := bufio.NewReader(bytes.NewReader(buf[:n]))

		req, err := http.ReadRequest(reader)

		if err != nil {
			fmt.Println("error parsing request string to http request object ", err)
		} else if req.Method == "POST" {
			fmt.Println("received", req.Method, req.Host, req.Body)

			bodyBytes, err := io.ReadAll(req.Body)

			if err != nil {
				fmt.Println("error reading request body ", err)
			}

			// if the request is to the domain abdul.com then we will create a new hostname
			if req.Host == "abdul.com" {
				s.createHostName(string(bodyBytes), &conn, req)
			}
		}

		// fmt.Println("received", string(buf[:n]))
	}
}

func (s *Server) createHostName(requestedName string, conn *net.Conn, req *http.Request) {
	// check if the requested name is already in use
	// if it is then return an error

	// if it is not then create a new hostname and return the hostname
	s.lockConn.Lock()
	if s.connections[requestedName] == nil {
		s.connections[requestedName] = conn

		// send the response back to the client
		resp := "HTTP/1.1 200 OK\r\nContent-Length: " + strconv.Itoa(len(requestedName)) + "\r\n\r\n" + requestedName
		(*conn).Write([]byte(resp))
	} else {
		// send the response back to the client
		resp := "HTTP/1.1 400 Bad Request\r\nContent-Length: " + strconv.Itoa(len("Hostname already in use")) + "\r\n\r\n" + "Hostname already in use"
		(*conn).Write([]byte(resp))
	}
	s.lockConn.Unlock()

}

func (s *Server) removeTheConnection(connection *net.Conn) {
	s.lockConn.Lock()
	for key, value := range s.connections {
		if value == connection {
			delete(s.connections, key)
		}
	}
	s.lockConn.Unlock()
}

func StartNewServer(listenAddress string) {
	server := Server{
		ListenAddress: listenAddress,
		connections:   make(map[string]*net.Conn),
	}
	server.startServer()
}
