package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Server struct {
	connections map[string]*net.TCPConn
}

func (s *Server) handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Client disconnected ", err)
				delete(s.connections, conn.RemoteAddr().String())
				break
			}
		}

		if strings.HasPrefix(string(buf[:n]), "domain") {
			parts := strings.SplitN(string(buf[:n]), "domain", 2)
			if len(parts) > 1 {
				contentFollowingDomain := parts[1]

				contentFollowingDomain = strings.TrimSpace(contentFollowingDomain)
				// check if the domain is already in the map
				exists := s.connections[contentFollowingDomain]
				if exists != nil {
					// we will forward the message to the other client
					_, err := conn.Write([]byte("false"))
					if err != nil {
						fmt.Println("Error writing to client: ", err)
					}
				} else {
					s.connections[contentFollowingDomain] = conn
				}

				// trim the content
				fmt.Println("Content following domain:", contentFollowingDomain)
				// Now contentFollowingDomain contains the contents following the word "domain"
			}
		}

		fmt.Println("Received from client: ", string(buf[:n]))
		fmt.Println("Number of connections: ", len(s.connections))

		// if len(s.connections) > 1 {
		// 	// we will forward the message to the other client
		// 	for _, c := range s.connections {
		// 		if c != conn {
		// 			_, err := c.Write(buf[:n])
		// 			// lets use copy instead of Write
		// 			// _, err := io.Copy(c, conn)
		// 			if err != nil {
		// 				fmt.Println("Error writing to client: ", err)
		// 			}
		// 		}
		// 	}
		// }
	}
}

func (s *Server) start() {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:8888")
	if err != nil {
		// handle error
		log.Fatal("Error resolving TCP address: ", err)
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		// handle error
		log.Fatal("Error listening on TCP: ", err)
	}

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Err accepting connection from client: ", err)
				continue
			}
		}
		// s.connections[conn.RemoteAddr().String()] = conn
		fmt.Println("New connection from: ", conn.RemoteAddr().String())

		go s.handleConnection(conn)
	}
}

func StartNewServer(address string) {
	server := &Server{
		connections: make(map[string]*net.TCPConn),
	}
	server.start()
}
